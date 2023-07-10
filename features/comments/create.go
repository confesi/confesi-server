package comments

// todo: make `ancestors` a single uint ID
// todo: db migrations to remove unused things and add new fields

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// func doesIdentifierExist(tx *gorm.DB, UID string, postId uint, identifier *uint) (error, db.CommentIdentifier, bool) {
// 	// check if user has already commented on this post with the same matchings
// 	possilbeIdentifier := db.CommentIdentifier{}

// 	query := tx.
// 		Where("user_id = ?", UID).
// 		Where("post_id = ?", postId).
// 		Where("identifier = ?", identifier)

// 	err := query.First(&possilbeIdentifier).Error
// 	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
// 		return serverError, possilbeIdentifier, false
// 	} else if errors.Is(err, gorm.ErrRecordNotFound) {
// 		// we have to create it and link the comment to this new one's ID
// 		return nil, possilbeIdentifier, false
// 	} else {
// 		// we link the comment to the ID of the existing one
// 		return nil, possilbeIdentifier, true
// 	}
// }

func getNextIdentifier(tx *gorm.DB, postId uint) (error, uint, uint) {
	highestIdentifier := db.Comment{}
	err := tx.
		Where("post_id = ?", postId).
		Order("identifier ASC").
		Find(&highestIdentifier).
		Limit(1).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return serverError, 0, 0
	}
	if errors.Is(err, gorm.ErrRecordNotFound) || highestIdentifier.Identifier == nil {
		return nil, 1, 0
	} else {

		return nil, *highestIdentifier.NumericalUser + 1, *highestIdentifier.NumericalUser
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
			fmt.Println("ROLL BACK")

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
			Where("comments.id = ?", req.ParentCommentID).
			Where("comments.post_id = ?", req.PostID).
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
		comment.Ancestors = pq.Int64Array{*req.ParentCommentID}
	} else {
		comment.Ancestors = pq.Int64Array{}
	}

	var nextIdentifier uint
	// is OP?
	if !isOp {
		err, nextIdentifier, _ = getNextIdentifier(tx, req.PostID)
		if err != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		}
	}

	if parentComment.NumericalUserIsOp != nil && *parentComment.NumericalUserIsOp {
		comment.NumericalReplyingUserIsOp = parentComment.NumericalUserIsOp
	} else {
		comment.NumericalReplyingUser = parentComment.NumericalUser
	}

	if isOp {
		t := true
		comment.NumericalUserIsOp = &t
	} else {
		comment.NumericalUser = &nextIdentifier
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
		fmt.Println("ROLL BACK")

		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	response.New(http.StatusCreated).Send(c)
}
