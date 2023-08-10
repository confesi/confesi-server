package posts

import (
	"confesi/db"
	"confesi/lib/response"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type sentimentAnalysis struct {
	Positive float64 `json:"positive"`
	Negative float64 `json:"negative"`
	Neutral  float64 `json:"neutral"`
	Compound float64 `json:"compound"`
}

func (h *handler) sentimentAnaylsis(c *gin.Context) {
	postID := c.Query("id")
	var post db.Post
	err := h.db.
		Preload("School").
		Preload("Category").
		Preload("Faculty").
		Preload("YearOfStudy").
		First(&post, postID).Error
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

	// sentiment analysis

	analysis := AnalyzeText(post.Title + "\n" + post.Content)

	// if all goes well, send status 200
	response.New(http.StatusOK).Val(analysis).Send(c)
	return
}
