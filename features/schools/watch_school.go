package schools

import (
	"confesi/db"
	"confesi/lib/masking"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

func (h *handler) watchSchool(c *gin.Context, token *auth.Token, req validation.WatchSchool, unmaskedId uint) error {
	school := db.SchoolFollow{
		UserID:   token.UID,
		SchoolID: unmaskedId,
	}
	err := h.DB.Create(&school).Error
	if err != nil {
		var pgErr *pgconn.PgError
		// Gorm doesn't properly handle duplicate errors: https://github.com/go-gorm/gorm/issues/4037
		if ok := errors.As(err, &pgErr); !ok {
			// if it's not a PostgreSQL error, return a generic server error
			return serverError
		}
		switch pgErr.Code {
		case "23505": // duplicate key value violates unique constraint
			return nil // just let the user know it's been watched, if it's already there
		case "23503": // foreign key constraint violation
			return invalidId // aka, you provided an school id to try watching
		default:
			// some other postgreSQL error
			return serverError
		}
	}
	return nil
}

func (h *handler) handleWatchSchool(c *gin.Context) {

	// extract request
	var req validation.WatchSchool

	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	unmaskedId, err := masking.Unmask(req.SchoolID)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	err = h.watchSchool(c, token, req, unmaskedId)
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
	response.New(http.StatusCreated).Send(c)
}
