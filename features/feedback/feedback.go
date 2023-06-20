package feedback

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleFeedback(c *gin.Context) {

	//Validate request
	var req validation.FeedbackDetails
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	// get the user's token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	// create a new feedback
	feedback := db.Feedback{
		UserID:  token.UID,
		Content: req.Message, //TODO: add type to DATABASE and add it to this struct
	}

	err = h.db.Create(&feedback).Error
	if err != nil {
		response.New(http.StatusCreated).Val(err).Send(c)
		return
	}

	// if all goes well, send 201
	response.New(http.StatusCreated).Send(c)
}
