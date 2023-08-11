package admin

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *handler) handleGetReportById(c *gin.Context) {
	// get id from query param id
	id := c.Query("id")

	idNumeric, err := strconv.Atoi(id)
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid id").Send(c)
		return
	}

	userRoles, err := getUserRoles(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		return
	}

	report := db.Report{}
	query := h.db.
		Preload("ReportType"). // preload the ReportType field of the Report
		Where("id = ?", idNumeric)

	if len(userRoles.SchoolMods) > 0 {
		query.Where("school_id IN ?", userRoles.SchoolMods)
	}

	err = query.First(&report).
		Error

	// If user does not have access to report or report does not exist, return 404
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.New(http.StatusNotFound).Err(notFound.Error()).Send(c)
		} else {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		}
		return
	}

	// if the post_id is not nil, preload the Post and its internal fields
	if report.PostID != nil {
		post := db.Post{}
		err := h.db.
			Preload("YearOfStudy").
			Preload("School").
			Preload("Category").
			Preload("Faculty").
			Where("id = ?", *report.PostID).
			First(&post).
			Error
		if err != nil {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
		report.Post = &post
		if !utils.ProfanityEnabled(c) {
			post = post.CensorPost()
		}
	}

	if report.CommentID != nil {
		comment := db.Comment{}
		err := h.db.
			Where("id = ?", *report.CommentID).
			First(&comment).
			Error
		if err != nil {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
		report.Comment = &comment
		if !utils.ProfanityEnabled(c) {
			comment = comment.CensorComment()
		}
	}

	response.New(http.StatusOK).Val(report).Send(c)
}
