package auth

import (
	"confesi/db"
	"confesi/lib/email"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// todo: add field like "you can only change your email once ever 90 days" in table or smth to keep track of it and enforce the restriction?

func (h *handler) handleUpdateEmail(c *gin.Context) {
	// let user know it won't update their home uni automatically (bug -> feature)

	// extract request body
	var req validation.UpdateEmail
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	// get the user's token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// extract domain from user's email
	domain, err := validation.ExtractEmailDomain(req.Email)
	if err != nil {
		response.New(http.StatusBadRequest).Err("error extracting domain from email").Send(c)
		return
	}

	// check if user's email is valid
	var school db.School
	err = h.db.Select("id").Where("domain = ?", domain).First(&school).Error
	if err != nil {
		response.New(http.StatusBadRequest).Err("domain doesn't belong to school").Send(c)
		return
	}

	// is the new email already in use?
	_, err = h.fb.AuthClient.GetUserByEmail(c, req.Email)
	if err == nil {
		// aka, user exists
		response.New(http.StatusBadRequest).Err("user already exists with this email").Send(c)
		return
	}

	// get the user's current email
	userEmail := token.Claims["email"].(string)

	// generate an email verificiation link
	link, err := h.fb.AuthClient.EmailVerificationLink(c, req.Email)
	em, err := email.New().
		To([]string{userEmail}, []string{}).
		Subject("Confesi Email Verification").
		LoadVerifyEmailTemplate(link)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	res, err := em.Send()
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	fmt.Println(res)
	response.New(http.StatusOK).Send(c)
}
