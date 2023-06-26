package admin

import (
	"confesi/db"
	"confesi/lib/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleFeedbackID(c *gin.Context) {
	feedbackID := c.Param("id")

	_, err := strconv.ParseInt(feedbackID, 10, 64)
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid feedback id").Send(c)
		return
	}
	feedback := db.Feedback{}
	err = h.db.Model(&db.Feedback{}).Where("id = ?", feedbackID).First(&feedback).Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
	}

	// if all goes well, send 200
	response.New(http.StatusOK).Val(feedback).Send(c)
	return
}
