package notifications

import (
	"confesi/db"
	"confesi/lib/fire"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"context"
	"fmt"
	"net/http"

	"firebase.google.com/go/messaging"
	"github.com/gin-gonic/gin"
)

// todo: updates if exists with time, else removes?
// todo: cron job to remove "dead" tokens
func (h *handler) handleSetToken(c *gin.Context) {

	// validate request
	var req validation.FcmTokenQuery
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	// Validate FCM token
	if !isValidFcmToken(h.fb.MsgClient, req.Token) {
		response.New(http.StatusBadRequest).Err("invalid fcm token").Send(c)
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	topics := []db.Topic{}

	// start a transaction
	tx := h.db.Begin()

	// if something goes ary, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err("server error").Send(c)
			return
		}
	}()

	err = tx.
		Find(&topics, "user_id = ?", token.UID).
		Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	failedSubs := []string{}
	for i := range topics {
		fire.SubToTopics(c, h.fb.MsgClient, req.Token, []string{topics[i].Name})
	}

	notification := db.Notification{
		UserID: token.UID,
		Token:  req.Token,
	}

	err = h.db.
		FirstOrCreate(&notification, notification).
		Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err(failedSubbing.Error()).Val(failedSubs).Send(c)
		return
	}

	if len(failedSubs) > 0 {
		response.New(http.StatusInternalServerError).Err(failedSubbing.Error()).Val(failedSubs).Send(c)
		return
	}

	response.New(http.StatusOK).Send(c)
}

func isValidFcmToken(client *messaging.Client, token string) bool {
	message := &messaging.Message{
		Token: token,
	}

	response, err := client.SendDryRun(context.Background(), message)
	if err != nil {
		// Handle error
		return false
	}

	fmt.Println(response)

	return true
}
