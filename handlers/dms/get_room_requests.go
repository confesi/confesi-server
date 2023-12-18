package dms

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleGetRoomRequests(c *gin.Context) {
	// Authenticate the user and obtain their token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("failed to get user token").Send(c)
		return
	}

	// Fetch the user along with their RoomRequests
	var user db.User
	err = h.db.
		Table("users").
		Where("id = ?", token.UID).
		First(&user).
		Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err("failed to fetch user").Send(c)
		return
	}

	// Respond with the RoomRequests of the user
	response.New(http.StatusOK).Val(user.RoomRequests).Send(c)
}
