package auth

import (
	"confesi/db"
	"confesi/lib/email"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
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

	// check if user's email is valid
	var school db.School
	err = h.db.Select("id").Where("domain = ?", domain).First(&school).Error
	if err != nil {
		response.New(http.StatusBadRequest).Err("domain doesn't belong to school").Send(c)
		return
	}

	// check if user's faculty is valid (aka, the faculty exists in the database)
	var faculty db.Faculty
	err = h.db.Select("id").Where("faculty = ?", req.Faculty).First(&faculty).Error
	if err != nil {
		response.New(http.StatusBadRequest).Err("faculty doesn't exist").Send(c)
		return
	}

	// new user
	newUser := (&auth.UserToCreate{}).
		Email(req.Email).
		Password(req.Password).
		Disabled(false)

	firebaseUser, err := h.fb.AuthClient.CreateUser(c, newUser)
	if err != nil {
		if strings.Contains(err.Error(), "EMAIL_EXISTS") {
			response.New(http.StatusConflict).Err("email already exists").Send(c)
		} else {
			response.New(http.StatusInternalServerError).Err("server error").Send(c)
		}
		return
	}

	verificationEmailSent := true
	err = email.SendVerificationEmail(c, h.fb.AuthClient, req.Email)
	if err != nil {
		verificationEmailSent = false
	}

	user := db.User{
		ID:          firebaseUser.UID,
		SchoolID:    school.ID,
		YearOfStudy: req.YearOfStudy,
		FacultyID:   uint(faculty.ID),
		ModID:       db.ModEnableID, // everyone starts off okay, but if they get sus... they'll get their account nerfed pretty quickly
	}

	// save user to postgres
	err = h.db.Create(&user).Error
	// we don't catch this error, because it will just show itself in the user's token as "sync: false" or DNE

	// on success of both user being created in firebase and postgres, change their token to "double verified"
	err = h.fb.AuthClient.SetCustomUserClaims(c, firebaseUser.UID, map[string]interface{}{
		"sync":  true,
		"roles": []string{}, //! default users have no roles, VERY IMPORTANT
	})
	// we don't catch this error, because it will just show itself in the user's token as "sync: false" or DNE

	// send response
	response.New(http.StatusCreated).Val(map[string]bool{"verification_sent": verificationEmailSent}).Send(c)
}
