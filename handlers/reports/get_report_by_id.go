package reports

import (
	"confesi/db"
	"confesi/lib/encryption"
	"confesi/lib/response"
	"confesi/lib/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *handler) handleGetReportById(c *gin.Context) {
	// get the user's token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// get id from query param id
	id := c.Query("id")
	unmaskedId, err := encryption.Unmask(id)
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid id").Send(c)
		return
	}

	report := db.Report{}

	err = h.db.
		Preload("ReportType"). // preload the ReportType field of the Report
		Where("id = ? AND reported_by = ?", unmaskedId, token.UID).
		First(&report).
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
	if report.PostID != nil {
		post := db.Post{}
		err := h.db.
			Preload("YearOfStudy").
			Preload("School").  // Preload the User field of the Post
			Preload("Faculty"). // Preload the User field of the Post
			Preload("Category").
			Where("id = ?", *report.PostID).
			First(&post).
			Error
		if err != nil {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
		report.Post = &post

		if post.Hidden {
			report.Post = nil
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

		if comment.Hidden {
			report.Comment = nil
		}
	}

	response.New(http.StatusOK).Val(report).Send(c)
}
