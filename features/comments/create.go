package comments

import (
	"confesi/config"
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func doesIdentifierExist(tx *gorm.DB, UID string, postId uint, identifier *uint, parentIdentifier *uint) (error, db.CommentIdentifier, bool) {

	// check if user has already commented on this post with same matchings
	possilbeIdentifier := db.CommentIdentifier{}

	err := tx.
		Where("user_id = ?", UID).
		Where("post_id = ?", postId).
		Where("identifier = ?", identifier).
		Where("parent_identifier = ?", parentIdentifier).
		First(&possilbeIdentifier).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return serverError, possilbeIdentifier, false
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// we have to create it and link comment to this new one's ID
		return nil, possilbeIdentifier, false
	} else {
		// we link comment to the ID of the existing one
		return nil, possilbeIdentifier, true
	}
}

func getNextIdentifier(tx *gorm.DB, postId uint) (error, uint) {
	highestIdentifier := db.CommentIdentifier{}
	err := tx.
		Where("post_id = ?", postId).
		Order("identifier ASC").
		Find(&highestIdentifier).
		Limit(1).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return serverError, 0
	}
	if errors.Is(err, gorm.ErrRecordNotFound) || highestIdentifier.Identifier == nil {
		return nil, 1
	} else {

		return nil, *highestIdentifier.Identifier + 1
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

	if req.ParentCommentID != nil {
		// is parent comment

		// get parent comment
		parentCommentIdentifier := db.CommentIdentifier{}
		parentComment := db.Comment{}

		// they are trying to create a threaded comment
		err = tx.
			Model(&parentComment).
			Joins("JOIN comment_identifiers ON comments.identifier_id = comment_identifiers.id").
			Where("comments.id = ?", req.ParentCommentID).
			Where("comments.post_id = ?", req.PostID).
			UpdateColumns(map[string]interface{}{
				"children_count": gorm.Expr("children_count + ?", 1),
			}).
			First(&parentCommentIdentifier).
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
		if len(parentComment.Ancestors) > config.MaxCommentThreadDepthExcludingRoot-1 {
			// can't thread comments this deep
			tx.Rollback()
			response.New(http.StatusBadRequest).Err(threadDepthError.Error()).Send(c)
			return
		}
		comment.Ancestors = append(parentComment.Ancestors, *req.ParentCommentID)

		var nextIdentifier uint

		// is OP?
		if !isOp {
			err, nextIdentifier = getNextIdentifier(tx, req.PostID)
			if err != nil {
				tx.Rollback()
				response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			}
		}

		var identifierIDForComment uint
		var newlyInsertedCommentIdentifier db.CommentIdentifier

		// is there already an existing identifier like this?
		err, possibleIdentifier, exists := doesIdentifierExist(tx, token.UID, req.PostID, &nextIdentifier, parentCommentIdentifier.Identifier)
		if err != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		} else if exists {
			identifierIDForComment = possibleIdentifier.ID
		} else {
			newlyInsertedCommentIdentifier = db.CommentIdentifier{
				UserID:           token.UID,
				PostID:           req.PostID,
				Identifier:       &nextIdentifier,
				ParentIdentifier: parentCommentIdentifier.Identifier,
			}
			err = tx.
				Create(&newlyInsertedCommentIdentifier).
				Error
			if err != nil {
				tx.Rollback()
				response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
				return
			}
			identifierIDForComment = newlyInsertedCommentIdentifier.ID
		}

	} else {
		// is not parent comment
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
