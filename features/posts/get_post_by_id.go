package posts

import (
	"confesi/db"
	"confesi/lib/response"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *handler) handleGetPostById(c *gin.Context) {
	postID := c.Query("id")
	var post db.Post
	err := h.db.Preload("School").Preload("Faculty").First(&post, postID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.New(http.StatusBadRequest).Err("post not found").Send(c)
			return
		}
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	if post.Hidden {
		response.New(http.StatusGone).Err("post removed").Send(c)
		return
	}
	response.New(http.StatusOK).Val(post).Send(c)
	return
}
