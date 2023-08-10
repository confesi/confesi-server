package feedback

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *handler) handleFeedback(c *gin.Context) {

	// validate request
	var req validation.FeedbackDetails
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	// get the user's token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// match the req.Type to the feedback_type table
	var feedbackType db.FeedbackType
	err = h.db.Where("type = ?", req.Type).First(&feedbackType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.New(http.StatusBadRequest).Err(feedbackTypeDoesntExist.Error()).Send(c)
			return
		}
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// create new feedback
	feedback := db.Feedback{
		UserID:  token.UID,
		Content: req.Message,
		TypeID:  feedbackType.ID.Val,
	}

	err = h.db.Create(&feedback).Error
	if err != nil {
		response.New(http.StatusCreated).Err(serverError.Error()).Send(c)
		return
	}

	// if all goes well, send 201
	response.New(http.StatusCreated).Send(c)
}
