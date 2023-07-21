package auth

import (
	"confesi/lib/email"
	"confesi/lib/response"
	"confesi/lib/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// todo: add custom rate limiting to this? or via redis?

var (
	errorSendingEmail = errors.New("error sending email")
)

func (h *handler) handleResendEmailVerification(c *gin.Context) {
	// get the user's token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// get the user's current email from their token
	userEmail := token.Claims["email"].(string)

	// resend the verification email
	err = email.SendVerificationEmail(c, h.fb.AuthClient, userEmail)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(errorSendingEmail.Error()).Send(c)
	}

	response.New(http.StatusOK).Send(c)
}
