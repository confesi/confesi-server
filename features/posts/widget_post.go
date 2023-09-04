package posts

import (
	"confesi/db"
	"confesi/lib/response"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type widgetPost struct {
	// just the ID, title, body, and school img
	ID           db.EncryptedID `json:"id"`
	Title        string         `json:"title"`
	Content      string         `json:"content"`
	SchoolImgUrl string         `json:"school_img_url"`
}

func (h *handler) handleGetWidgetPost(c *gin.Context) {

	var post db.Post

	err := h.db.
		Preload("School").
		Where("hidden = ?", false).
		Where("DATE(created_at) = (SELECT DATE(created_at) FROM posts WHERE hidden = ? ORDER BY created_at DESC LIMIT 1)", false).
		Order("trending_score DESC").
		First(&post).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.New(http.StatusNotFound).Err("post not found").Send(c)
			return
		}
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	post = post.CensorPost()

	widgetPost := widgetPost{
		ID:           post.ID,
		Title:        post.Title,
		Content:      post.Content,
		SchoolImgUrl: post.School.ImgUrl,
	}

	response.New(http.StatusOK).Val(widgetPost).Send(c)
	return
}
