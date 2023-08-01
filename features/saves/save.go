package saves

import (
	"confesi/db"
	"confesi/lib/logger"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
	"fmt"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
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
		logger.StdErr(errors.New(fmt.Sprintf("invalid content type: %s", req.ContentType)), nil, nil, nil, nil)
		return serverError
	}
	if err != nil {
		var pgErr *pgconn.PgError
		// Gorm doesn't properly handle duplicate errors: https://github.com/go-gorm/gorm/issues/4037
		if ok := errors.As(err, &pgErr); !ok {
			// if it's not a PostgreSQL error, return a generic server error
			return serverError
		}
		switch pgErr.Code {
		case "23505": // duplicate key value violates unique constraint
			return nil // just let the user know it's been created, if it's already there
		case "23503": // foreign key constraint violation
			return invalidId // aka, you provided an invalid post/comment id to try saving
		default:
			// some other postgreSQL error
			return serverError
		}
	}
	return nil
}

func (h *handler) handleSave(c *gin.Context) {
	// extract request
	var req validation.SaveContentDetails

	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	err = h.saveContent(c, token, req)
	if err != nil {
		// switch over err
		switch err {
		case invalidId:
			response.New(http.StatusBadRequest).Err(err.Error()).Send(c)
		default:
			response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		}
		return
	}

	// if all goes well, send 201
	response.New(http.StatusCreated).Send(c)
}
