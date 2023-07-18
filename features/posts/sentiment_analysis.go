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

type sentimentAnalysis struct {
	Positive float64 `json:"positive"`
	Negative float64 `json:"negative"`
	Neutral  float64 `json:"neutral"`
	Compound float64 `json:"compound"`
}

// SentimentAnaylsis godoc
//
//	@Summary		Sentiment Anaylsis.
//	@Description	Sentiment anaylsis for posts by ID.
//	@Tags			Posts
//	@Accept			application/json
//	@Produce		application/json
//	@Security		BearerAuth
//	@Security		X-AppCheck-Token
//	@Param			id	query		string					true	"Example: 27"
//	@Success		200	{object}	docs.SentimentAnaylsis	"Sentiment Anaylsis"
//	@Failure		400	{object}	docs.PostNotFound		"Post Not Found"
//	@Failure		410	{object}	docs.PostRemoved		"Post Removed"
//	@Failure		500	{object}	docs.ServerError		"Server Error"
//
//	@Router			/posts/sentiment [get]
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
	parsedtext := sentitext.Parse(post.Title+"\n"+post.Content, lexicon.DefaultLexicon)
	sentiment := sentitext.PolarityScore(parsedtext)

	analysis := sentimentAnalysis{
		Positive: sentiment.Positive,
		Negative: sentiment.Negative,
		Neutral:  sentiment.Neutral,
		Compound: sentiment.Compound,
	}

	// if all goes well, send status 200
	response.New(http.StatusOK).Val(analysis).Send(c)
	return
}
