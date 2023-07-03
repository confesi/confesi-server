package comments

import (
	"confesi/config"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleGetReplies(c *gin.Context) {
	// extract request
	var req validation.RepliesCommentQuery
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	var commentDetails []CommentDetail

	next := time.UnixMilli(int64(req.Next))

	err = h.db.
		Preload("Identifier").
		Table("comments").
		Where("ancestors[1] = ?", req.ParentComment).
		Where("created_at < ?", next).
		Order("created_at DESC").
		Limit(config.RepliesLoadedManually).
		Find(&commentDetails).
		Error

	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// if all good, send 200
	response.New(http.StatusOK).Val(commentDetails).Send(c)
}
