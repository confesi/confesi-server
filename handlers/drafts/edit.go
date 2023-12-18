package drafts

import (
	"confesi/db"
	"confesi/lib/encryption"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleEditDraft(c *gin.Context) {
	// validate the json body from request
	var req validation.EditDraft
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

	unmaskedId, err := encryption.Unmask(req.DraftID)
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid id").Send(c)
		return
	}

	updates := map[string]interface{}{
		"title":   req.Title,
		"content": req.Body,
	}

	results := h.db.Model(&db.Draft{}).
		Where("id = ?", unmaskedId).
		Where("user_id = ?", token.UID).
		Updates(updates)

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
