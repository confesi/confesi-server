package comments

import (
	"confesi/config"
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

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

	comment := db.Comment{
		UserID:  token.UID,
		PostID:  req.PostID,
		Content: req.Content,
	}

	// they are trying to create a threaded comment
	parentComment := db.Comment{}
	if req.ParentCommentID != nil {
		err = tx.Where("id = ?", req.ParentCommentID).
			First(&parentComment).
			UpdateColumn("children_count", gorm.Expr("children_count + ?", 1)).
			Error
		if err != nil {
			// parent comment not found
			if errors.Is(err, gorm.ErrRecordNotFound) {
				tx.Rollback()
				response.New(http.StatusBadRequest).Err("parent comment doesn't exist").Send(c)
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
	} else {
		// its a root comment
		comment.Ancestors = pq.Int64Array{}
	}

	// check if there already exists a comment identifier
	commentIdentifier := db.CommentIdentifier{}
	err = tx.
		Where("user_id = ?", token.UID).
		Where("post_id = ?", req.PostID).
		First(&commentIdentifier).
		Error

	// check if identifier record already exists
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		// if not, create a new one
		var post db.Post
		err = tx.
			Where("id = ?", req.PostID).
			First(&post).
			Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.New(http.StatusBadRequest).Err("referenced post not found").Send(c)
			}
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			tx.Rollback()
			return
		}
		if post.UserID == token.UID {
			// user is OP
			newOpCommentIdentifier := db.CommentIdentifier{
				UserID: token.UID,
				PostID: req.PostID,
				IsOp:   true,
			}
			err = tx.Create(&newOpCommentIdentifier).Error
			if err != nil {
				tx.Rollback()
				response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
				return
			}
			comment.IdentifierID = newOpCommentIdentifier.ID
		} else {
			// user is not OP
			// list all the already existing comment identifiers and get the one with the highest "identifier" column, then save one with that + 1
			var highestIdentifierSoFar db.CommentIdentifier
			err = tx.
				Where("post_id = ?", req.PostID).
				Order("identifier ASC").
				Find(&highestIdentifierSoFar).
				Limit(1).
				Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				tx.Rollback()
				response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
				return
			}
			var newIdentifier int64
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newIdentifier = 1
			} else {
				if highestIdentifierSoFar.Identifier == nil {
					newIdentifier = 1
				} else {
					newIdentifier = *highestIdentifierSoFar.Identifier + 1
				}
			}
			// save new comment identifier
			newNotOpCommentIdentifier := db.CommentIdentifier{
				UserID:     token.UID,
				PostID:     req.PostID,
				Identifier: &newIdentifier,
				IsOp:       false,
			}
			err = tx.Create(&newNotOpCommentIdentifier).
				Error
			if err != nil {
				tx.Rollback()
				response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
				return
			}
			comment.IdentifierID = newNotOpCommentIdentifier.ID
		}
	} else {
		// if it already exists, set it for the soon-to-be-created comment
		comment.IdentifierID = commentIdentifier.ID
	}

	// save the comment
	err = tx.Create(&comment).Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// if all goes well, respond with a 201 & commit the transaction
	tx.Commit()
	response.New(http.StatusCreated).Send(c)
}
