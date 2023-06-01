package auth

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/validation"
	"net/http"
	"strings"
	"time"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// TODO: add email verification, and route to enable checking if email is verified to pass through the middleware
// Example creating a Firebase user
func (h *handler) handleRegister(c *gin.Context) {

	// extract request body
	var req validation.CreateAccountDetails

	// create a binding instance with the validator, check if json valid, if so, deserialize into req
	binding := &validation.DefaultBinding{
		Validator: validator.New(),
	}
	if err := binding.Bind(c.Request, &req); err != nil {
		response.New(http.StatusBadRequest).Err("invalid json").Send(c)
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
		ID:        firebaseUser.UID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     req.Email,
		// TODO: replace with non-dummy values:
		YearOfStudy: 3,
		FacultyID:   1,
		SchoolID:    1,
		ModID:       1,
	}

	// save user to postgres
	err = h.db.Create(&user).Error
	if err != nil {
		// If firebase account creation succeeds, but postgres profile save fails
		response.New(http.StatusCreated).Val("auth").Send(c)
		return
	}

	// on success of both user being created in firebase and postgres, change their token to "double verified"
	claims := map[string]interface{}{
		"profile_created": true,
	}

	// Set the custom token on the user
	err = h.fb.AuthClient.SetCustomUserClaims(c, firebaseUser.UID, claims)
	if err != nil {
		// If firebase account creation succeeds, but postgres profile save fails
		response.New(http.StatusCreated).Val("auth").Send(c)
		return
	}

	// if this succeeds, send back success to indicate the user should reload their account because both their account & profile
	// has been created
	response.New(http.StatusCreated).Val("full").Send(c)
}
