package posts

import (
	"confesi/db"
	"confesi/lib/response"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/grassmudhorses/vader-go/lexicon"
	"github.com/grassmudhorses/vader-go/sentitext"
	"gorm.io/gorm"
)

func (h *handler) sentimentAnaylsis(c *gin.Context) {
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

	// sentiment analysis
	parsedtext := sentitext.Parse(post.Content, lexicon.DefaultLexicon)
	sentiment := sentitext.PolarityScore(parsedtext)

	// Generate JSON response
	json := gin.H{"Positive": sentiment.Positive, "Negative": sentiment.Negative, "Neutral": sentiment.Neutral, "Compound": sentiment.Compound}

	response.New(http.StatusOK).Val(json).Send(c)
	return
}
