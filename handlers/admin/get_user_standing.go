package admin

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

type userStandingResult struct {
	Limited bool `json:"limited"`
	Banned  bool `json:"banned"`
}

func (h *handler) handleGetUserStanding(c *gin.Context) {
	// validate request
	var req validation.UserQuery
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	user := db.User{}
	err = h.db.Select("is_limited").Model(&db.User{}).Where("id = ?", req.UserID).First(&user).Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	// get the user's disabled firebase status
	fbUser, err := h.fb.AuthClient.GetUser(c, req.UserID)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	// if all goes well, send 200
	response.New(http.StatusOK).Val(userStandingResult{
		Limited: user.IsLimited,
		Banned:  fbUser.Disabled,
	}).Send(c)
}
