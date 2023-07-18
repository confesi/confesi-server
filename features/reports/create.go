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

	// start a transaction
	tx := h.db.Begin()
	// if something goes ary, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}()

	var modelMatcher interface{}
	report := db.Report{}
	if req.ContentType == "post" {
		report.PostID = &req.ContentID
		modelMatcher = &db.Post{}
	} else if req.ContentType == "comment" {
		report.CommentID = &req.ContentID
		modelMatcher = &db.Comment{}
	} else {
		// should never happen... but to be defensive
		tx.Rollback()
		response.New(http.StatusBadRequest).Err(invalidContentId.Error()).Send(c)
		return
	}

	// inc the report count for the post/comment by 1
	err = tx.
		Model(&modelMatcher).
		Where("id = ?", req.ContentID).
		Update("report_count", gorm.Expr("report_count + 1")).
		Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// match the req.Type to the report_type table
	var reportType db.ReportType
	err = tx.Where("type = ?", req.Type).First(&reportType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			response.New(http.StatusBadRequest).Err(reportTypeDoesntExist.Error()).Send(c)
			return
		}
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	report.ReportedBy = token.UID
	report.Description = req.Description
	report.TypeID = uint(reportType.ID)

	err = tx.Create(&report).Error
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
			tx.Rollback()
			response.New(http.StatusConflict).Err(reportAlreadyExists.Error()).Send(c)
			return
		case "23503": // foreign key constraint violation
			tx.Rollback()
			response.New(http.StatusBadRequest).Err(invalidContentId.Error()).Send(c)
			return // aka, you provided an invalid post/comment id to try saving
		default:
			// some other postgreSQL error
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	// commit the transaction
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// if all goes well, send 201
	response.New(http.StatusCreated).Send(c)
}
