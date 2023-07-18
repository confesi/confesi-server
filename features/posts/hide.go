package posts

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HidePosts godoc
//
//	@Summary		Hide Post.
//	@Description	Hide posts based on ID.
//	@Tags			Posts
//	@Accept			application/json
//	@Produce		application/json
//	@Security		BearerAuth
//	@Security		X-AppCheck-Token
//	@Param			Body	body		string						true	"Json Example"	SchemaExample({\n"post_id": 150\n})
//	@Success		200		{object}	docs.PostHidden				"Post Hidden"
//	@Failure		400		{object}	docs.NoPostFoundWithThisID	"No Post Found With This ID"
//	@Failure		500		{object}	docs.ServerError			"Server Error"
//
//	@Router			/posts/hide [patch]
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

	// Update the "hidden" field on a post.
	result := h.db.
		Model(&db.Post{}).
		Where("id = ? AND user_id = ?", req.PostID, token.UID).
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
