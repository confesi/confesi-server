package admin

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func (h *handler) handleUserStanding(c *gin.Context) {

	//Validate request
	var req validation.UserStanding
	binding := &validation.DefaultBinding{
		Validator: validator.New(),
	}
	if err := binding.Bind(c.Request, &req); err != nil {
		fmt.Println(err)
		response.New(http.StatusBadRequest).Err(fmt.Sprintf("failed validation: %v", err)).Send(c)
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error 1").Send(c)
		return
	}

	var stance int
	switch req.Standing {
	case "enabled":
		stance = 1
	case "limited":
		stance = 2
	case "banned":
		stance = 3
	default:
		//! Should Never Get Here
		response.New(http.StatusBadRequest).Err("invalid standing").Send(c)
		return
	}

	// update the user's standing to the new standing

	err = h.db.Model(&db.User{}).Where("id = ?", token.UID).Update("ModID", stance).Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error 2").Send(c)
		return
	}

	// if all goes well, send 200
	response.New(http.StatusOK).Send(c)
}
