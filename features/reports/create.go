package reports

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func (h *handler) handleCreateReport(c *gin.Context) {

	// validate request
	var req validation.ReportQuery
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	// get the user's token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	report := db.Report{}
	if req.ContentType == "post" {
		report.PostID = &req.ContentID
	} else if req.ContentType == "comment" {
		report.CommentID = &req.ContentID
	} else {
		// should never happen... but to be defensive
		response.New(http.StatusBadRequest).Err("invalid content type").Send(c)
		return
	}

	// match the req.Type to the report_type table
	var reportType db.ReportType
	err = h.db.Where("type = ?", req.Type).First(&reportType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.New(http.StatusBadRequest).Err(reportTypeDoesntExist.Error()).Send(c)
			return
		}
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	report.ReportedBy = token.UID
	report.Description = req.Description
	report.TypeID = uint(reportType.ID)

	err = h.db.Create(&report).Error
	if err != nil {
		var pgErr *pgconn.PgError
		// Gorm doesn't properly handle duplicate errors: https://github.com/go-gorm/gorm/issues/4037
		if ok := errors.As(err, &pgErr); !ok {
			// if it's not a PostgreSQL error, return a generic server error
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
		switch pgErr.Code {
		case "23505": // duplicate key value violates unique constraint
			response.New(http.StatusConflict).Err(reportAlreadyExists.Error()).Send(c)
			return
		case "23503": // foreign key constraint violation
			response.New(http.StatusBadRequest).Err(invalidContentId.Error()).Send(c)
			return // aka, you provided an invalid post/comment id to try saving
		default:
			// some other postgreSQL error
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	// if all goes well, send 201
	response.New(http.StatusCreated).Send(c)
}
