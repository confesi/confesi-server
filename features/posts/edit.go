package posts

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleEditPost(c *gin.Context) {
	// validate the json body from request
	var req validation.EditPost
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

	updates := map[string]interface{}{
		"edited":  true,
		"title":   req.Title,
		"content": req.Body,
	}

	// Update the `Title`/`Body` and `Edited` fields of the comment in a single query
	results := h.db.Model(&db.Post{}).
		Where("id = ?", req.PostID).
		Where("hidden = false").
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
