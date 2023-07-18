package admin

import (
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleReviewContentByMod(c *gin.Context) {
	// validate request
	var req validation.UpdateReviewedByModQuery
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	var table string
	if req.ContentType == "comment" {
		table = "comments"
	} else if req.ContentType == "post" {
		table = "posts"
	} else {
		response.New(http.StatusBadRequest).Err(invalidValue.Error()).Send(c)
		return
	}

	err = h.db.
		Table(table).
		Where("id = ?", req.ContentID).
		Update("reviewed_by_mod", req.ReviewedByMod).
		Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	response.New(http.StatusOK).Send(c)
}
