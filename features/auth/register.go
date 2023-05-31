package auth

// TODO: initializer syntax

import (
	"confesi/config"
	"confesi/db"
	"confesi/lib/response"
	"net/http"
	"regexp"
	"strings"
	"time"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

type request struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func validEmail(email string) (bool, error) {
	// check email length
	if len(email) > config.MaxEmailLength || len(email) < config.MinEmailLength {
		return false, nil
	}
	// check email format
	pattern := `(?i)^([a-z0-9_+]([a-z0-9_+.]*[a-z0-9_+])?)@([a-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,6})`
	input := []byte(email)
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}
	return regex.Match(input), nil
}

func validPassword(password string) bool {
	// check password length
	if len(password) > config.MaxPasswordLength || len(password) < config.MinPasswordLength {
		return false
	}
	return true
}

func domainFromEmail(email string) (string, error) {
	pattern := `\@[A-Za-z0-9]+\.[A-Za-z]{2,6}`
	input := []byte(email)
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	return regex.FindString(string(input)), nil
}

// TODO: add email verification, and route to enable checking if email is verified to pass through the middleware
// Example creating a Firebase user
func (h *handler) handleRegister(c *gin.Context) {

	// deserialize request
	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		response.New(http.StatusBadRequest).Err("invalid json").Send(c)
		return
	}

	// check if email is valid
	if valid, err := validEmail(req.Email); !valid {
		response.New(http.StatusBadRequest).Err("invalid email").Send(c)
		return
	} else if err != nil {
		response.New(http.StatusBadRequest).Err("error validating email").Send(c)
		return
	}

	// check pw meets standards
	if !validPassword(req.Password) {
		response.New(http.StatusBadRequest).Err("invalid password").Send(c)
		return
	}

	// extract domain from user's email
	domain, err := domainFromEmail(req.Email)
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
