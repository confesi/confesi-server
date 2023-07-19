package admin

import (
	"confesi/db"
	"confesi/lib/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FeedbackByID godoc
//
//	@Summary		Feedback By ID.
//	@Description	Obtain specific feedback by its ID
//	@Tags			Admin
//	@Accept			application/json
//	@Produce		application/json
//	@Security		BearerAuth
//	@Security		X-AppCheck-Token
//	@Param			feedback_id	path		int					true	"Feedback ID"
//	@Success		200			{object}	docs.FeedbackByID	"Feedback Result"
//	@Failure		500			{object}	docs.ServerError	"Server Error"
//
//	@Router			/admin/feedback/{feedback_id} [get]
func (h *handler) handleFeedbackID(c *gin.Context) {
	feedbackID := c.Param("feedbackID")

	_, err := strconv.ParseInt(feedbackID, 10, 64)
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid feedback id").Send(c)
		return
	}
	feedback := db.Feedback{}
	err = h.db.Model(&db.Feedback{}).Where("id = ?", feedbackID).First(&feedback).Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	// if all goes well, send 200
	response.New(http.StatusOK).Val(feedback).Send(c)
}
