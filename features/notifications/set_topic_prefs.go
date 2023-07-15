package notifications

import (
	"confesi/db"
	fcm "confesi/lib/firebase_cloud_messaging"
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
	fcm.SendSyncNotification(*h.db, h.fb.MsgClient, token.UID, fcm.SyncTypeNotificationPrefs)
	response.New(http.StatusOK).Send(c)
}
