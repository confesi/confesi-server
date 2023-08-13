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

	// first or create the FcmTopicPref record for the user
	topicPref := db.FcmTopicPref{
		UserID: token.UID,
	}

	// perform a FirstOrCreate to check if the record exists
	err = h.db.
		Where(&db.FcmTopicPref{UserID: token.UID}).
		Attrs(&topicPref).
		FirstOrCreate(&topicPref).
		Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// check if any of the fields have changed
	changed := false
	if topicPref.DailyHottest != *req.DailyHottest ||
		topicPref.Trending != *req.Trending ||
		topicPref.RepliesToYourComments != *req.RepliesToYourComments ||
		topicPref.CommentsOnYourPosts != *req.CommentsOnYourPosts ||
		topicPref.VotesOnYourComments != *req.VotesOnYourComments ||
		topicPref.VotesOnYourPosts != *req.VotesOnYourPosts ||
		topicPref.QuotesOfYourPosts != *req.QuotesOfYourPosts {
		changed = true
	}

	// update the prefs with the new values
	topicPref.DailyHottest = *req.DailyHottest
	topicPref.Trending = *req.Trending
	topicPref.RepliesToYourComments = *req.RepliesToYourComments
	topicPref.CommentsOnYourPosts = *req.CommentsOnYourPosts
	topicPref.VotesOnYourComments = *req.VotesOnYourComments
	topicPref.VotesOnYourPosts = *req.VotesOnYourPosts
	topicPref.QuotesOfYourPosts = *req.QuotesOfYourPosts

	// save the record (create or update)
	err = h.db.Save(&topicPref).Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// if nothing changed, don't send a sync request
	if !changed {
		response.New(http.StatusOK).Send(c)
		return
	}

	// don't catch errors, just hope it works; the user can manually sync if needed
	var tokens []string

	// fetch user's tokens from the database
	err = h.db.Table("users").
		Select("fcm_tokens.token").
		Joins("JOIN fcm_tokens ON fcm_tokens.user_id = users.id").
		Where("users.id = ?", token.UID).
		Pluck("fcm_tokens.token", &tokens).
		Error

	if err == nil && len(tokens) > 0 {
		go fcm.New(h.fb.MsgClient).
			ToTokens(tokens).
			WithData(builders.NotificationSettingsSyncData()).
			Send()
	} else if err != nil {
		// "handle" the error if fetching tokens fails
		logger.StdInfo(fmt.Sprintf("failed to send sync request for set topic prefs: %v", err))
	}

	response.New(http.StatusOK).Send(c)
}
