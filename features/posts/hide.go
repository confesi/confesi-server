package posts

import (
	"confesi/db"
	"confesi/lib/encryption"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleHidePost(c *gin.Context) {
	// validate request json
	var req validation.HidePost
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

	unmaskedId, err := encryption.Unmask(req.PostID)
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid id").Send(c)
		return
	}

	// Update the "hidden" field on a post.
	result := h.db.
		Model(&db.Post{}).
		Where("id = ? AND user_id = ?", unmaskedId, token.UID).
		Update("hidden", "true")

	if result.Error != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	if result.RowsAffected == 0 {
		response.New(http.StatusBadRequest).Err("no post found with this ID").Send(c)
		return
	}

	response.New(http.StatusOK).Send(c)
	return
}
