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

func (h *handler) handleGetYourComments(c *gin.Context) {
	// extract request
	var req validation.YourCommentsQuery
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	next := time.UnixMicro(int64(req.Next))

	commentDetails := []CommentDetail{}
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
			WHERE created_at < ?
			AND user_id = ?
			ORDER BY created_at DESC
			LIMIT ?
		`, token.UID, next, token.UID, config.YourCommentsPageSize).
		Find(&commentDetails).
		Error

	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	for i := range commentDetails {
		// create ref to comment
		comment := &commentDetails[i]
		if comment.Hidden {
			comment.Comment.Content = "[removed]"
			comment.Comment.Identifier = nil
		}
		// check if user is owner
		if comment.UserID == token.UID {
			comment.Owner = true
		}
	}

	response.New(http.StatusOK).Val(commentDetails).Send(c)
}
