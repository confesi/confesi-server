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
