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

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	var commentDetails []CommentDetail

	next := time.UnixMilli(int64(req.Next))

	err = h.db.
		Preload("Identifier").
		Raw(`
			SELECT comments.*, 
				COALESCE(
					(SELECT votes.vote
					FROM votes
					WHERE votes.comment_id = comments.id
						AND votes.user_id = ?
					LIMIT 1),
					'0'::vote_score_value
				) AS user_vote
			FROM comments
			WHERE ancestors[1] = ?
				AND created_at < ?
			ORDER BY created_at DESC
			LIMIT ?
		`, token.UID, req.ParentComment, next, config.RepliesLoadedManually).
		First(&commentDetails).
		Error

	for i := range commentDetails {
		// create reference to comment
		comment := &commentDetails[i]
		if comment.Hidden {
			comment.Comment.Content = "[removed]"
			comment.Comment.Identifier = nil
		}
	}

	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// if all good, send 200
	response.New(http.StatusOK).Val(commentDetails).Send(c)
}
