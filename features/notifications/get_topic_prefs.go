package notifications

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleGetTopicPrefs(c *gin.Context) {
	// validate request
	var req validation.FcmNotifictionPref
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	topicPrefs := db.FcmTopicPref{}
	err = h.db.
		First(&topicPrefs, "user_id = ?", token.UID).
		Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	response.New(http.StatusOK).Val(topicPrefs).Send(c)
}
