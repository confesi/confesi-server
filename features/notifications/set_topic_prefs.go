package notifications

import (
	"confesi/config/builders"
	"confesi/db"
	fcm "confesi/lib/firebase_cloud_messaging"
	"confesi/lib/logger"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"fmt"
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

	topicPrefs := map[string]interface{}{
		"daily_hottest":            req.DailyHottest,
		"trending":                 req.Trending,
		"replies_to_your_comments": req.RepliesToYourComments,
		"comments_on_your_posts":   req.CommentsOnYourPosts,
		"votes_on_your_comments":   req.VotesOnYourComments,
		"votes_on_your_posts":      req.VotesOnYourPosts,
		"quotes_of_your_posts":     req.QuotesOfYourPosts,
	}

	err = h.db.
		Model(&db.FcmTopicPref{}).
		Where("user_id = ?", token.UID).
		Updates(topicPrefs).
		Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// don't catch errors, just hope it works, else, the user can manually sync
	var tokens []string

	// Fetch user's tokens from the database
	err = h.db.Table("users").
		Select("fcm_tokens.token").
		Joins("JOIN fcm_tokens ON fcm_tokens.user_id = users.id").
		Where("users.id = ?", token.UID).
		Pluck("fcm_tokens.token", &tokens).
		Error

	fmt.Println(tokens)

	if err == nil && len(tokens) > 0 {
		fcm.New(h.fb.MsgClient).
			ToTokens(tokens).
			WithData(builders.NotificationSettingsSyncData()).
			Send(*h.db)

	} else {
		// handle the error if fetching tokens fails
		logger.StdInfo(fmt.Sprintf("failed to send sync request for set topic prefs: %v", err))
	}

	response.New(http.StatusOK).Send(c)
}
