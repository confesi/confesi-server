package notifications

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleSetTopicPrefs(c *gin.Context) {
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

	topicPrefs := db.FcmTopicPref{
		UserID:          token.UID,
		DailyHottest:    req.DailyHottest,
		TrendingAll:     req.TrendingAll,
		TrendingHome:    req.TrendingHome,
		TrendingWatched: req.TrendingWatched,
		NewFeatures:     req.NewFeatures,
	}
	err = h.db.
		Save(&topicPrefs).
		Where("user_id = ?", token.UID).
		Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	response.New(http.StatusOK).Val(topicPrefs).Send(c)

	// todo: send FCM message to "sync" topic settings to all this user_id's devices
}
