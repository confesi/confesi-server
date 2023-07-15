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
		return nil, true, *comment.Numerics.NumericalUser
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
	if errors.Is(err, gorm.ErrRecordNotFound) || highestIdentifier.Numerics.NumericalUser == nil {
		return nil, 1
	} else {
		return nil, *highestIdentifier.Numerics.NumericalUser + 1
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

	if parentComment.Numerics.NumericalUserIsOp {
		comment.Numerics.NumericalReplyingUserIsOp = true
	} else {
		comment.Numerics.NumericalReplyingUserIsOp = false
		comment.Numerics.NumericalReplyingUser = parentComment.Numerics.NumericalUser
	}

	if isOp {
		comment.Numerics.NumericalUserIsOp = true
	} else {
		comment.Numerics.NumericalUserIsOp = false
		err, alreadyPosted, userNumeric := getAlreadyPostedNumericalUser(tx, req.PostID, token.UID)
		if err != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
		if alreadyPosted {
			comment.Numerics.NumericalUser = &userNumeric
		} else {
			comment.Numerics.NumericalUser = &nextIdentifier
		}

	}

	// create the comment
	err = tx.Create(&comment).
		Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// if all goes well, respond with a 201 & commit the transaction
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// to-send-to tokens
	var tokens []string

	// post owner
	err = h.db.
		Table("fcm_tokens").
		Select("fcm_tokens.token").
		Joins("JOIN users ON users.id = fcm_tokens.user_id").
		Joins("JOIN posts ON posts.user_id = users.id").
		Where("posts.id = ?", req.PostID).
		Pluck("fcm_tokens.token", &tokens).
		Error

	if err == nil {
		fcm.New(h.fb.MsgClient).
			ToTokens(tokens).
			WithMsg(builders.CommentAddedToPost(req.Content)).
			Send(*h.db)
	}

	// if threaded comment, parent comment
	if req.ParentCommentID != nil {
		err = h.db.
			Table("fcm_tokens").
			Select("fcm_tokens.token").
			Joins("JOIN users ON users.id = fcm_tokens.user_id").
			Joins("JOIN comments ON comments.user_id = users.id").
			Where("comments.id = ?", req.ParentCommentID).
			Pluck("fcm_tokens.token", &tokens).
			Error
		if err == nil {
			fcm.New(h.fb.MsgClient).
				ToTokens(tokens).
				WithMsg(builders.ThreadedCommentReply(req.Content)).
				Send(*h.db)
		}

	}

	response.New(http.StatusCreated).Send(c)
}
