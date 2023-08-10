package admin

import (
	"confesi/db"
	"confesi/lib/masking"
	"confesi/lib/response"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *handler) handleFeedbackID(c *gin.Context) {
	feedbackID := c.Param("feedbackID")

	unmaskedFeedbackId, err := masking.Unmask(feedbackID)
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid feedback id").Send(c)
		return
	}

	feedback := db.Feedback{}
	err = h.db.Model(&db.Feedback{}).Where("id = ?", unmaskedFeedbackId).First(&feedback).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		response.New(http.StatusBadRequest).Err("feedback not found").Send(c)
		return
	}

	// if all goes well, send 200
	response.New(http.StatusOK).Val(feedback).Send(c)
	return
}
