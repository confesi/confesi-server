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
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
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

	var firebaseUser *auth.UserRecord
	userToSaveToPostgres := db.User{}
	var userIdForPostgres string
	// ensure the token is valid, aka, there is some valid user
	if req.AlreadyExistingAccToken != "" {

		fmt.Println("has already existing acc. token")

		token, err := h.fb.AuthClient.VerifyIDToken(c, req.AlreadyExistingAccToken)
		if err != nil {
			response.New(http.StatusBadRequest).Err("invalid existing user token").Send(c)
			return
		}

		// check if this user has already been registered by email
		_, err = h.fb.AuthClient.GetUserByEmail(c, req.Email)
		if err == nil {
			tx.Rollback()
			response.New(http.StatusBadRequest).Err("account already upgraded").Send(c)
			return
		}

		// get firebase account by this UID
		_, err = h.fb.AuthClient.GetUser(c, token.UID)
		if err != nil {
			tx.Rollback()
			response.New(http.StatusBadRequest).Err("invalid already existing account UID").Send(c)
			return
		}
		// check if found user is anonymous
		if token.Claims["provider_id"] != "anonymous" {
			tx.Rollback()
			response.New(http.StatusBadRequest).Err("already existing account is not anonymous").Send(c)
			return
		}

		// new firebase user
		userToUpdate := (&auth.UserToUpdate{}).
			Email(req.Email).
			Password(req.Password).
			Disabled(false)

		_, err = h.fb.AuthClient.UpdateUser(c, token.UID, userToUpdate)
		userIdForPostgres = token.UID
	} else {
		firebaseUser, err = h.fb.AuthClient.CreateUser(c, newUser)
		if err == nil {
			userIdForPostgres = firebaseUser.UID
		}
	}

	if err != nil {
		if strings.Contains(err.Error(), "EMAIL_EXISTS") {
			tx.Rollback()
			response.New(http.StatusConflict).Err("email already exists").Send(c)
		} else {
			fmt.Println(err)
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		}
		return
	}

	userToSaveToPostgres.SchoolID = school.ID
	userToSaveToPostgres.ID = userIdForPostgres

	// save user to postgres
	err = h.db.Create(&userToSaveToPostgres).Error
	// we don't catch this error, because it will just show itself in the user's token as "sync: false" or DNE

	// on success of both user being created in firebase and postgres, change their token to "double verified" via the "sync" field
	h.fb.AuthClient.SetCustomUserClaims(c, userIdForPostgres, map[string]interface{}{
		"sync":  true,
		"roles": []string{}, //! default users have no roles, VERY IMPORTANT
	})
	// we don't catch this error, because it will just show itself in the user's token as "sync: false" or DNE

	// commit the transaction
	err = tx.Commit().Error
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// send response & don't care if email sends
	go email.SendVerificationEmail(c, h.fb.AuthClient, req.Email)
	response.New(http.StatusCreated).Send(c)
}
