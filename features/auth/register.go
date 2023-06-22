package auth

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"
	"strings"
	"time"

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

	user := db.User{
		ID: firebaseUser.UID,
		// UTC times in milliseconds of when the user was created and last updated (both "default" to when user was created initially)
		CreatedAt:   time.UnixMilli(firebaseUser.UserMetadata.CreationTimestamp),
		UpdatedAt:   time.UnixMilli(firebaseUser.UserMetadata.CreationTimestamp),
		Email:       req.Email,
		SchoolID:    school.ID,
		YearOfStudy: req.YearOfStudy,
		FacultyID:   uint(faculty.ID),
		ModID:       db.ModEnableID, // everyone starts off okay, but if they get sus... that's another story
	}

	// save user to postgres
	err = h.db.Create(&user).Error
	if err != nil {
		// If firebase account creation succeeds, but postgres profile save fails
		response.New(http.StatusCreated).Val("auth").Send(c)
		return
	}

	// on success of both user being created in firebase and postgres, change their token to "double verified"
	err = h.fb.AuthClient.SetCustomUserClaims(c, firebaseUser.UID, map[string]interface{}{
		"profile_created": true,
		"roles":           []string{}, // default users have no roles
	})
	if err != nil {
		// If firebase account creation succeeds, but postgres profile save fails
		response.New(http.StatusCreated).Val("auth").Send(c)
		return
	}

	// if this succeeds, send back success to indicate the user should reload their account because both their account & profile
	// has been created
	response.New(http.StatusCreated).Val("full").Send(c)
}
