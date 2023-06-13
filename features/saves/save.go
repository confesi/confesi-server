package saves

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/validation"
	"errors"
	"fmt"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	ServerError = "server error"
)

func (h *handler) saveContent(c *gin.Context, token *auth.Token, req validation.SaveContentDetails) error {
	// bit of saved content
	var err error
	if req.ContentType == "post" {
		savedPost := db.SavedPost{
			UserID: token.UID,
			PostID: req.ContentID,
		}
		err = h.db.Create(&savedPost).Error
	} else if req.ContentType == "comment" {
		savedComment := db.SavedComment{
			UserID:    token.UID,
			CommentID: req.ContentID,
		}
		err = h.db.Create(&savedComment).Error
	} else {
		// should never happen with validated struct, but to be defensive
		return errors.New(ServerError)
	}
	if err != nil {
		var pgErr *pgconn.PgError
		// Gorm doesn't properly handle duplicate errors: https://github.com/go-gorm/gorm/issues/4037
		if ok := errors.As(err, &pgErr); !ok || pgErr.Code != "23505" {
			return errors.New(ServerError)
		}
	}
	return nil
}

func (h *handler) handleSave(c *gin.Context) {
	// extract request
	var req validation.SaveContentDetails

	// create validator
	validator := validator.New()

	// create a binding instance with the validator, check if json valid, if so, deserialize into req
	binding := &validation.DefaultBinding{
		Validator: validator,
	}
	if err := binding.Bind(c.Request, &req); err != nil {
		response.New(http.StatusBadRequest).Err(fmt.Sprintf("failed validation: %v", err)).Send(c)
		return
	}

	// TODO: START: once firebase user utils function is merged, use that instead to be cleaner
	// get firebase user
	user, ok := c.Get("user")
	if !ok {
		response.New(http.StatusInternalServerError).Err(ServerError).Send(c)
		return
	}

	token, ok := user.(*auth.Token)
	if !ok {
		response.New(http.StatusInternalServerError).Err(ServerError).Send(c)
		return
	}
	// TODO: END

	err := h.saveContent(c, token, req)
	if err != nil {
		// all returned errors are just general client-facing "server errors"
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		return
	}

	// if all goes well, send 201
	response.New(http.StatusCreated).Send(c)
}
