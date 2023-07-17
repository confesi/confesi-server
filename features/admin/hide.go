package admin

import (
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleHideContent(c *gin.Context) {

	// validate request json
	var req validation.HideContent
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	var table string
	if req.ContentType != "comment" && req.ContentType != "post" {
		response.New(http.StatusBadRequest).Err(invalidValue.Error()).Send(c)
		return
	}
	table = req.ContentType + "s"

	// Update the "hidden" field on a comment.
	result := h.db.
		Table(table).
		Where("id = ?", req.ContentID).
		Update("hidden", req.Hide)

	if result.Error != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	if result.RowsAffected == 0 {
		response.New(http.StatusBadRequest).Err("no content found with this ID").Send(c)
		return
	}

	response.New(http.StatusOK).Send(c)
	return
}
