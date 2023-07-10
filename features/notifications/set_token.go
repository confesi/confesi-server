package notifications

import (
	"confesi/lib/utils"
	"confesi/lib/validation"

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

	// token, err := utils.UserTokenFromContext(c)
	// if err != nil {
	// 	response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
	// 	return
	// }
}
