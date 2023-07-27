package auth

import (
	"confesi/lib/email"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleSendPasswordResetEmail(c *gin.Context) {

	// extract request body
	var req validation.EmailQuery
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	// resend the verification email
	err = email.SendPasswordResetEmail(c, h.fb.AuthClient, req.Email)
	if err != nil && !errors.Is(err, email.ErrorUserDoesntExist) {
		response.New(http.StatusInternalServerError).Err(errorSendingEmail.Error()).Send(c)
		return
	} else if errors.Is(err, email.ErrorUserDoesntExist) {
		response.New(http.StatusBadRequest).Err(email.ErrorUserDoesntExist.Error()).Send(c)
		return
	}

	response.New(http.StatusOK).Send(c)
}
