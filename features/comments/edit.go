package comments

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleEditComment(c *gin.Context) {
	// validate the json body from request
	var req validation.EditComment
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

	// Update the `Content` and `Edited` fields of the comment in a single query
	results := h.db.Model(&db.Comment{}).
		Where("id = ?", req.CommentID).
		Where("hidden = false").
		Where("user_id = ?", token.UID).
		Updates(map[string]interface{}{
			"content": req.Content,
			"edited":  true,
		})

	if results.RowsAffected == 0 {
		response.New(http.StatusNotFound).Err(notFound.Error()).Send(c)
		return
	}

	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	response.New(http.StatusOK).Send(c)
}
