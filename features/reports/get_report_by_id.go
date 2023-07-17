package reports

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

type fetchedReport struct {
	Report         db.Report `json:"report"`
	ContentRemoved bool      `json:"content_removed"`
}

func (h *handler) handleGetReportById(c *gin.Context) {
	// get id from query param id
	id := c.Query("id")

	// get the user's token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	fetchedReport := fetchedReport{}

	idNumeric, err := strconv.Atoi(id)
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid id").Send(c)
		return
	}

	err = h.db.
		Preload("ReportType"). // preload the ReportType field of the Report
		Where("id = ? AND reported_by = ?", idNumeric, token.UID).
		First(&fetchedReport.Report).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.New(http.StatusNotFound).Err(invalidContentId.Error()).Send(c)
		} else {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		}
		return
	}

	// if the post_id is not nil, preload the Post and its internal fields
	if fetchedReport.Report.PostID != nil {
		post := db.Post{}
		err := h.db.
			Preload("School").  // Preload the User field of the Post
			Preload("Faculty"). // Preload the User field of the Post
			Where("id = ?", *fetchedReport.Report.PostID).
			First(&post).
			Error
		if err != nil {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
		fetchedReport.Report.Post = &post

		if post.Hidden {
			fetchedReport.ContentRemoved = true
			fetchedReport.Report.Post = nil
		}
	}

	if fetchedReport.Report.CommentID != nil {
		comment := db.Comment{}
		err := h.db.
			Where("id = ?", *fetchedReport.Report.CommentID).
			First(&comment).
			Error
		if err != nil {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
		fetchedReport.Report.Comment = &comment

		if comment.Hidden {
			fetchedReport.ContentRemoved = true
			fetchedReport.Report.Comment = nil
		}
	}

	response.New(http.StatusOK).Val(fetchedReport).Send(c)
}
