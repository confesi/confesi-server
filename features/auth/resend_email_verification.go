package auth

import (
	"confesi/lib/email"
	"confesi/lib/logger"
	"confesi/lib/response"
	"confesi/lib/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// todo: add custom rate limiting to this? or via redis?

func (h *handler) handleResendEmailVerification(c *gin.Context) {
	// get the user's token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// get the user's current email from their token
	userEmail := token.Claims["email"].(string)

	// if user already verified, ignore
	if token.Claims["email_verified"].(bool) {
		response.New(http.StatusBadRequest).Val("already verified").Send(c)
		return
	}

	// resend the verification email
	err = email.SendVerificationEmail(c, h.fb.AuthClient, userEmail)
	if err != nil && !errors.Is(err, email.ErrorNoLinkGeneratedError) {
		logger.StdErr(err, nil, nil, nil, nil)
		response.New(http.StatusInternalServerError).Err(errorSendingEmail.Error()).Send(c)
		return
	} else if errors.Is(err, email.ErrorNoLinkGeneratedError) {
		logger.StdErr(err, nil, nil, nil, nil)
		response.New(http.StatusBadRequest).Err(email.ErrorNoLinkGeneratedError.Error()).Send(c)
		return
	}

	response.New(http.StatusOK).Send(c)
}
