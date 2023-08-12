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

	unmaskedId, err := encryption.Unmask(req.DraftID)
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid id").Send(c)
		return
	}

	results := h.db.
		Where("id = ?", unmaskedId).
		Where("user_id = ?", token.UID).
		Delete(&db.Draft{})

	if results.Error != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	response.New(http.StatusOK).Send(c)
}
