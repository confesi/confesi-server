package comments

import (
	"confesi/config/builders"
	"confesi/db"
	fcm "confesi/lib/firebase_cloud_messaging"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// (error, bool, uint) -> (error, alreadyPosted, numericalUser)
func getAlreadyPostedNumericalUser(tx *gorm.DB, postID uint, userID string) (error, bool, uint) {
	comment := db.Comment{}
	err := tx.
		Where("user_id = ?", userID).
		Where("post_id = ?", postID).
		Where("numerical_user IS NOT NULL").
		First(&comment).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return serverError, false, 0
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, 0
	} else {
		return nil, true, *comment.NumericalUser
	}
}

func getNextIdentifier(tx *gorm.DB, postId uint) (error, uint) {
	highestIdentifier := db.Comment{}
	err := tx.
		Where("post_id = ?", postId).
		Order("numerical_user ASC").
		Find(&highestIdentifier).
		Limit(1).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return serverError, 0
	}
	if errors.Is(err, gorm.ErrRecordNotFound) || highestIdentifier.NumericalUser == nil {
		return nil, 1
	} else {
		return nil, *highestIdentifier.NumericalUser + 1
	}
}

func (h *handler) handleCreate(c *gin.Context) {

	// validate the json body from request
	var req validation.CreateComment
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	// get user token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// start a transaction
	tx := h.db.Begin()

	// if something goes ary, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err("server error").Send(c)
			return
		}
	}()

	var post db.Post
	err = tx.
		Where("id = ?", req.PostID).
		First(&post).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.New(http.StatusBadRequest).Err("post not found").Send(c)
			return
		}
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		tx.Rollback()
		return
	}

	isOp := post.UserID == token.UID

	// base comment
	comment := db.Comment{
		UserID:  token.UID,
		PostID:  req.PostID,
		Content: req.Content,
	}

	parentComment := db.Comment{}

	// they are trying to create a threaded comment
	if req.ParentCommentID != nil {

		// parent comment

		err = tx.
			Where("comments.id = ? AND comments.post_id = ?", req.ParentCommentID, req.PostID).
			Find(&parentComment).
			Updates(map[string]interface{}{
				"children_count": gorm.Expr("children_count + ?", 1),
			}).
			Error
		if err != nil {
			// parent comment not found
			if errors.Is(err, gorm.ErrRecordNotFound) {
				tx.Rollback()
				response.New(http.StatusBadRequest).Err("parent-comment and post combo doesn't exist").Send(c)
				return
			}
			// some other error
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
		if parentComment.ParentRoot == nil {
			comment.ParentRoot = &parentComment.ID
		} else {
			comment.ParentRoot = parentComment.ParentRoot
		}
	} else {

		comment.ParentRoot = nil
	}

	var nextIdentifier uint
	// is OP?
	if !isOp {
		err, nextIdentifier = getNextIdentifier(tx, req.PostID)
		if err != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	// to allow for pointers
	t := true
	f := false

	if parentComment.NumericalUserIsOp != nil && *parentComment.NumericalUserIsOp {

		comment.NumericalReplyingUserIsOp = &t
	} else {

		comment.NumericalReplyingUserIsOp = &f
		comment.NumericalReplyingUser = parentComment.NumericalUser
	}

	if isOp {

		comment.NumericalUserIsOp = &t
	} else {

		comment.NumericalUserIsOp = &f
		err, alreadyPosted, userNumeric := getAlreadyPostedNumericalUser(tx, req.PostID, token.UID)
		if err != nil {

			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}

		if alreadyPosted {
			comment.NumericalUser = &userNumeric
		} else {
			comment.NumericalUser = &nextIdentifier
		}

	}

	// create the comment
	err = tx.Create(&comment).
		Error
	if err != nil {
		var pgErr *pgconn.PgError
		// Gorm doesn't properly handle duplicate errors: https://github.com/go-gorm/gorm/issues/4037
		if ok := errors.As(err, &pgErr); !ok {
			// if it's not a PostgreSQL error, return a generic server error
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
		switch pgErr.Code {

		case "23503": // foreign key constraint violation
			tx.Rollback()
			response.New(http.StatusBadRequest).Err("parent comment doesn't exist").Send(c)
			return
		default:
			// some other postgreSQL error
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	if req.ParentCommentID != nil {
		res := tx.
			Model(&db.Post{}).
			Where("id = ?", req.PostID).
			Updates(map[string]interface{}{"comment_count": gorm.Expr("comment_count + ?", 1)})
		if res.Error != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	// if all goes well, respond with a 201 & commit the transaction
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// to-send-to postTokens
	var postTokens []string

	// post owner
	err = h.db.
		Table("fcm_tokens").
		Select("fcm_tokens.token").
		Joins("JOIN users ON users.id = fcm_tokens.user_id").
		Joins("JOIN posts ON posts.user_id = users.id").
		Where("posts.id = ? AND users.id <> ?", req.PostID, token.UID).
		Pluck("fcm_tokens.token", &postTokens).
		Error

	if err == nil && len(postTokens) > 0 {
		fcm.New(h.fb.MsgClient).
			ToTokens(postTokens).
			WithMsg(builders.CommentAddedToPostNoti(req.Content)).
			WithData(builders.CommentAddedToPostData(comment.ID, req.PostID)).
			Send(*h.db)
	}

	// respond "success" BEFORE sending FCM
	response.New(http.StatusCreated).Val(CommentDetail{Comment: comment, UserVote: 0, Owner: true}).Send(c)

	// if threaded comment, parent comment
	if req.ParentCommentID != nil {
		// to-send-to threadTokens
		var threadTokens []string
		err = h.db.
			Table("fcm_tokens").
			Select("fcm_tokens.token").
			Joins("JOIN users ON users.id = fcm_tokens.user_id").
			Joins("JOIN comments ON comments.user_id = users.id").
			Where("comments.id = ? AND users.id <> ?", req.ParentCommentID, token.UID).
			Pluck("fcm_tokens.token", &threadTokens).
			Error
		if err == nil && len(threadTokens) > 0 {
			fcm.New(h.fb.MsgClient).
				ToTokens(threadTokens).
				WithMsg(builders.ThreadedCommentReplyNoti(req.Content)).
				WithData(builders.ThreadedCommentReplyData(*req.ParentCommentID, comment.ID, req.PostID)).
				Send(*h.db)
		}
	}

}
