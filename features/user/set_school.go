package user

import (
	"confesi/db"
	"confesi/lib/masking"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

func (h *handler) handleSetSchool(c *gin.Context) {

	// validate the json body from request
	var req validation.UpdateSchool
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	unmaskedId, err := masking.Unmask(req.SchoolID)
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid id").Send(c)
		return
	}

	// update the user's school
	res := h.db.
		Model(&db.User{}).
		Where("id = ?", token.UID).
		Update("school_id", unmaskedId)
	if res.Error != nil {
		var pgErr *pgconn.PgError
		// Gorm doesn't properly handle some errors: https://github.com/go-gorm/gorm/issues/4037
		if ok := errors.As(res.Error, &pgErr); !ok {
			// if it's not a PostgreSQL error, return a generic server error
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
		switch pgErr.Code {
		case "23503": // foreign key constraint violation
			response.New(http.StatusBadRequest).Err("invalid school").Send(c)
			return

		default:
			// some other postgreSQL error
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return

		}
	}
	if res.RowsAffected == 0 {
		// no rows were affected, meaning the user doesn't exist
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// say 200 if all goes well
	response.New(http.StatusOK).Send(c)
}
