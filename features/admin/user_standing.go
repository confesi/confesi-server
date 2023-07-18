package admin

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleUserStanding(c *gin.Context) {

	//Validate request
	var req validation.UserStanding
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	// get the user's token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	// convert the standing to a stance variable number
	var stance int
	switch req.Standing {
	case "enabled":
		stance = 1
	case "limited":
		stance = 2
	case "banned":
		stance = 3
	default:
		// should never get here, but to be defensive
		response.New(http.StatusBadRequest).Err("invalid standing").Send(c)
		return
	}

	// update the user's standing to the new standing
	err = h.db.Model(&db.User{}).Where("id = ?", token.UID).Update("mod_id", stance).Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	// if all goes well, send 200
	response.New(http.StatusOK).Send(c)
}
