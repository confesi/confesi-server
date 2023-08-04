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

	// start a transaction
	tx := h.db.Begin()
	// if something goes ary, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}()

	topicPrefs := db.FcmTopicPref{}
	err = tx.
		First(&topicPrefs, "user_id = ?", token.UID).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// if nothing is found for them, create a new pref record, and return it
		topicPrefs = db.FcmTopicPref{
			UserID:                token.UID,
			DailyHottest:          true,
			Trending:              true,
			RepliesToYourComments: true,
			CommentsOnYourPosts:   true,
			VotesOnYourComments:   true,
			VotesOnYourPosts:      true,
			QuotesOfYourPosts:     true,
		}
		err = tx.Create(&topicPrefs).Error
		if err != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}
	// commit the transaction
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	response.New(http.StatusOK).Val(topicPrefs).Send(c)
}
