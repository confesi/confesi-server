package admin

import (
	"confesi/config/builders"
	fcm "confesi/lib/firebase_cloud_messaging"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleSendNotification(c *gin.Context) {

	// validate request
	var req validation.SendNotification
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	// obtain fcm tokens for affected users
	var tokens []string
	err = h.db.
		Table("fcm_tokens").
		Select("fcm_tokens.token").
		Joins("JOIN users ON users.id = fcm_tokens.user_id").
		Where("users.id IN ?", req.UserIDs).
		Pluck("fcm_tokens.token", &tokens).
		Error

	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	go fcm.New(h.fb.MsgClient).
		ToTokens(tokens).
		WithMsg(builders.AdminSendNotificationNoti(req.Title, req.Body)).
		WithData(req.Data).
		ShownInBackgroundOnly(req.Background).
		Send()

	// if all goes well send 200
	response.New(http.StatusOK).Send(c)
}

// todo: send fcm
