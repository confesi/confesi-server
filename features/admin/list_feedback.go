package admin

import (
	"confesi/config"
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

type feedbackFetchResults struct {
	Feedback []db.Feedback `json:"feedback"`
	Next     *int64        `json:"next"`
}

func (h *handler) handleListFeedback(c *gin.Context) {

	// extract request
	var req validation.FeedbackCursor
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	feedbackFetchResults := feedbackFetchResults{}

	err = h.db.
		Preload("Type").
		Where(req.Next.Cursor("created_at >")).
		Order("created_at ASC").
		Find(&feedbackFetchResults.Feedback).
		Limit(config.AdminFeedbackPageSize).
		Error

	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	if len(feedbackFetchResults.Feedback) > 0 {
		timeMicros := (feedbackFetchResults.Feedback[len(feedbackFetchResults.Feedback)-1].CreatedAt.Time).UnixMicro()
		feedbackFetchResults.Next = &timeMicros
	}

	response.New(http.StatusOK).Val(feedbackFetchResults).Send(c)
}
