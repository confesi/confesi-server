package auth

import (
	"confesi/db"
	"confesi/lib/email"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"fmt"
	"net/http"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

// TODO: add email verification, and route to enable checking if email is verified to pass through the middleware
// Example creating a Firebase user
func (h *handler) handleRegister(c *gin.Context) {

	// extract request body
	var req validation.CreateAccountDetails
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	// extract domain from user's email
	domain, err := validation.ExtractEmailDomain(req.Email)
	if err != nil {
		response.New(http.StatusBadRequest).Err("error extracting domain from email").Send(c)
		return
	}

	// start a transaction
	tx := h.db.Begin()
	// if something goes ary, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err("server error").Send(c)
			return
		}
	}()

	// check if user's email is valid
	var school db.School
	err = tx.Select("id").Where("domain = ?", domain).First(&school).Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusBadRequest).Err("domain doesn't belong to school").Send(c)
		return
	}

	// new firebase user
	newUser := (&auth.UserToCreate{}).
		Email(req.Email).
		Password(req.Password).
		Disabled(false)

	firebaseUser, err := h.fb.AuthClient.CreateUser(c, newUser)
	if err != nil {
		if strings.Contains(err.Error(), "EMAIL_EXISTS") {
			tx.Rollback()
			response.New(http.StatusConflict).Err("email already exists").Send(c)
		} else {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err("server error").Send(c)
		}
		return
	}

	verificationEmailSent := true
	err = email.SendVerificationEmail(c, h.fb.AuthClient, req.Email)
	if err != nil {
		fmt.Println("error, email not sent", err)
		verificationEmailSent = false
	}

	user := db.User{}
	user.ID = firebaseUser.UID
	user.SchoolID = school.ID

	// save user to postgres
	err = h.db.Create(&user).Error
	// we don't catch this error, because it will just show itself in the user's token as "sync: false" or DNE

	// on success of both user being created in firebase and postgres, change their token to "double verified"
	err = h.fb.AuthClient.SetCustomUserClaims(c, firebaseUser.UID, map[string]interface{}{
		"sync":  true,
		"roles": []string{}, //! default users have no roles, VERY IMPORTANT
	})
	// we don't catch this error, because it will just show itself in the user's token as "sync: false" or DNE

	// commit the transaction
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// send response
	response.New(http.StatusCreated).Val(map[string]bool{"verification_sent": verificationEmailSent}).Send(c)
}
