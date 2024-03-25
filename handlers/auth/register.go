package auth

import (
	"confesi/db"
	"confesi/lib/email"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
	"log/slog"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RegistrationError struct {
	PublicMessage string
	CustomCode    int
}

func (regErr *RegistrationError) Code() int {
	if regErr.CustomCode != 0 {
		return regErr.CustomCode
	}

	return http.StatusBadRequest
}

func (regErr *RegistrationError) Error() string {
	return regErr.PublicMessage
}

// Example creating a Firebase user
func (h *handler) handleRegister(c *gin.Context) {

	// extract request body
	var req validation.CreateAccountDetails
	err := utils.New(c).Validate(&req)
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid").Send(c)
		return
	}

	// extract domain from user's email
	domain, err := validation.ExtractEmailDomain(req.Email)
	if err != nil {
		response.New(http.StatusBadRequest).Err("error extracting domain from email").Send(c)
		return
	}

	// start a transaction
	err = h.db.Transaction(func(tx *gorm.DB) error {
		// check if user's email is valid
		var school db.School
		err := tx.Select("id").Where("domain = ?", domain).First(&school).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &RegistrationError{PublicMessage: "domain doesn't belong to school"}
			}
			return err
		}

		// new firebase user
		newUser := (&auth.UserToCreate{}).
			Email(req.Email).
			Password(req.Password).
			Disabled(false)

		var userIdForPostgres string
		var firebaseError error
		// ensure the token is valid, aka, there is some valid user
		if req.AlreadyExistingAccToken != "" {

			token, err := h.fb.AuthClient.VerifyIDToken(c, req.AlreadyExistingAccToken)
			if err != nil {
				return &RegistrationError{PublicMessage: "invalid existing user token"}
			}

			// check if this user has already been registered by email
			_, err = h.fb.AuthClient.GetUserByEmail(c, req.Email)
			if err == nil {
				return &RegistrationError{PublicMessage: "account already upgraded"}
			}

			// get firebase account by this UID
			_, err = h.fb.AuthClient.GetUser(c, token.UID)
			if err != nil {
				return &RegistrationError{PublicMessage: "invalid already existing account UID"}
			}
			// check if found user is anonymous
			if token.Claims["provider_id"] != "anonymous" {
				return &RegistrationError{PublicMessage: "already existing account is not anonymous"}
			}

			// new firebase user
			userToUpdate := (&auth.UserToUpdate{}).
				Email(req.Email).
				Password(req.Password).
				Disabled(false)

			_, firebaseError = h.fb.AuthClient.UpdateUser(c, token.UID, userToUpdate)
			userIdForPostgres = token.UID
		} else {
			var firebaseUser *auth.UserRecord
			firebaseUser, firebaseError = h.fb.AuthClient.CreateUser(c, newUser)
			if firebaseError == nil {
				userIdForPostgres = firebaseUser.UID
			}
		}

		if firebaseError != nil {
			if auth.IsEmailAlreadyExists(firebaseError) {
				return &RegistrationError{PublicMessage: "email already exists", CustomCode: http.StatusConflict}
			}
			return firebaseError
		}

		// save user to postgres
		err = h.db.Create(&db.User{
			SchoolID: school.ID,
			ID:       userIdForPostgres,
		}).Error
		// we don't catch this error, because it will just show itself in the user's token as "sync: false" or DNE
		if err != nil {
			slog.Error("Failed to save user to Postgres", "error", err)
		}

		// on success of both user being created in firebase and postgres, change their token to "double verified" via the "sync" field
		h.fb.AuthClient.SetCustomUserClaims(c, userIdForPostgres, map[string]interface{}{
			"sync":  true,
			"roles": []string{}, //! default users have no roles, VERY IMPORTANT
		})
		// we don't catch this error, because it will just show itself in the user's token as "sync: false" or DNE

		return nil
	})

	if err != nil {
		var regErr *RegistrationError
		if errors.As(err, &regErr) {
			response.New(regErr.Code()).Err(regErr.PublicMessage).Send(c)
			return
		}

		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// send response & don't care if email sends
	go func() {
		err := email.SendVerificationEmail(c, h.fb.AuthClient, req.Email)
		if err != nil {
			slog.Error("Failed to send verification e-mail in registration", "error", err)
		}
	}()
	response.New(http.StatusCreated).Send(c)
}
