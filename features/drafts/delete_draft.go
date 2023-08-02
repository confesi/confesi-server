package drafts

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleDeleteDraft(c *gin.Context) {
	// validate the json body from request
	var req validation.DeleteDraft
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	// get user token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// Update the `Title`/`Body` and `Edited` fields of the comment in a single query
	results := h.db.
		Where("id = ?", req.DraftID).
		Where("user_id = ?", token.UID).
		Delete(&db.Draft{})

	if results.Error != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	if results.RowsAffected == 0 {
		response.New(http.StatusNotFound).Err(notFound.Error()).Send(c)
		return
	}

	response.New(http.StatusOK).Send(c)
}
