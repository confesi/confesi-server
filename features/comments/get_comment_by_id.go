package comments

import (
	"confesi/db"
	"confesi/lib/response"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *handler) handleGetCommentById(c *gin.Context) {
	commentID := c.Query("id")
	var comment db.Comment
	err := h.db.Preload("Identifier").First(&comment, commentID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.New(http.StatusBadRequest).Err("comment not found").Send(c)
			return
		}
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	if comment.Hidden {
		response.New(http.StatusGone).Err("comment removed").Send(c)
		return
	}
	response.New(http.StatusOK).Val(comment).Send(c)
	return
}
