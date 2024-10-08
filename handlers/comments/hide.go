package comments

import (
	"confesi/db"
	"confesi/lib/encryption"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleHideComment(c *gin.Context) {

	// validate request json
	var req validation.HideComment
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

	unmaskedId, err := encryption.Unmask(req.CommentID)
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid id").Send(c)
		return
	}

	// Update the "hidden" field on a comment.
	result := h.db.
		Model(&db.Comment{}).
		Where("id = ? AND user_id = ?", unmaskedId, token.UID).
		Update("hidden", "true")

	if result.Error != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	if result.RowsAffected == 0 {
		response.New(http.StatusBadRequest).Err("no comment found with this ID").Send(c)
		return
	}

	response.New(http.StatusOK).Send(c)
	return

}
