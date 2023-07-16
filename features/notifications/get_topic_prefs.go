package notifications

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *handler) handleGetTopicPrefs(c *gin.Context) {

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	topicPrefs := db.FcmTopicPref{}
	err = h.db.
		First(&topicPrefs, "user_id = ?", token.UID).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.New(http.StatusBadRequest).Err("no entry found for user").Send(c)
		return
	}
	response.New(http.StatusOK).Val(topicPrefs).Send(c)
}
