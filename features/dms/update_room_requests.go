package dms

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleUpdateRoomRequests(c *gin.Context) {
	// validate request
	var req validation.UpdateRoomRequestable
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// update the user `RoomRequests` field with req.Requestable bool
	err = h.db.
		Model(&db.User{}).
		Where("id = ?", token.UID).
		Update("room_requests", req.Requestable).
		Error

	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	response.New(http.StatusOK).Send(c)
}
