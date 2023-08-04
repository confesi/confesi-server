package hideLog

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *handler) handleGetHideLogById(c *gin.Context) {
	// get id from query param id
	id := c.Query("id")

	// get the user's token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	idNumeric, err := strconv.Atoi(id)
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid id").Send(c)
		return
	}

	hideLog := db.HideLog{}

	err = h.db.
		Where("id = ? AND user_id = ?", idNumeric, token.UID).
		First(&hideLog).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.New(http.StatusNotFound).Err(nothingFoundForId.Error()).Send(c)
		} else {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		}
		return
	}

	// if the post_id is not nil, preload the Post and its internal fields
	if hideLog.PostID != nil {
		post := db.Post{}
		err := h.db.
			Preload("School").
			Preload("Category").
			Preload("Faculty").
			Preload("YearOfStudy").
			Where("id = ?", hideLog.PostID).
			First(&post).
			Error
		if err != nil {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
		hideLog.Post = &post

		if post.Hidden {
			hideLog.Post = nil
		}
	}

	if hideLog.CommentID != nil {
		comment := db.Comment{}
		err := h.db.
			Where("id = ?", hideLog.CommentID).
			First(&comment).
			Error
		if err != nil {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
		hideLog.Comment = &comment

		if comment.Hidden {
			hideLog.Comment = nil
		}
	}

	response.New(http.StatusOK).Val(hideLog).Send(c)
}
