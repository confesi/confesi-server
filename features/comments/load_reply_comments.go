package comments

import (
	"confesi/config"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReplyComments struct {
	Comments []CommentDetail `json:"comments"`
	Next     *int64          `json:"next"`
}

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

	fetchResults := ReplyComments{}

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
			`+req.Next.Cursor("AND created_at >")+`
			ORDER BY created_at ASC
			LIMIT ?
		`, token.UID, req.ParentComment, config.RepliesLoadedManually).
		Find(&fetchResults.Comments).
		Error

	if len(fetchResults.Comments) > 0 {
		timeMicros := (fetchResults.Comments[len(fetchResults.Comments)-1].CreatedAt.Time).UnixMicro()
		fetchResults.Next = &timeMicros
		for i := range fetchResults.Comments {
			// create reference to comment
			comment := &fetchResults.Comments[i]
			if comment.Hidden {
				comment.Comment.Content = "[removed]"
				comment.Comment.Identifier = nil
			}
			// check if user is owner
			if comment.UserID == token.UID {
				comment.Owner = true
			}
		}
	}

	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// if all good, send 200
	response.New(http.StatusOK).Val(fetchResults).Send(c)
}
