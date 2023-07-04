package saves

import (
	"confesi/features/comments"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"
	"time"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

type FetchedComments struct {
	Comments []comments.CommentDetail `json:"comments"`
	Next     *int64                   `json:"next"`
}

func (h *handler) getComments(c *gin.Context, token *auth.Token, req validation.SaveContentCursor) (*FetchedComments, error) {
	fetchResult := FetchedComments{}
	next := time.UnixMilli(int64(req.Next))

	query := `
		SELECT comments.*, saved_comments.updated_at,
			COALESCE(
				(
					SELECT votes.vote
					FROM votes
					WHERE votes.comment_id = comments.id
						AND votes.user_id = ?
					LIMIT 1
				),
				'0'::vote_score_value
			) AS user_vote
		FROM comments
		JOIN saved_comments ON comments.id = saved_comments.comment_id
		WHERE saved_comments.updated_at < ?
			AND saved_comments.user_id = ?
			AND comments.hidden = false
		ORDER BY saved_comments.updated_at DESC
		LIMIT ?
		`

	err := h.db.Raw(query, token.UID, next, token.UID, cursorSize).
		Preload("Identifier").
		Find(&fetchResult.Comments).Error

	if err != nil {
		return nil, err
	}

	if len(fetchResult.Comments) > 0 {
		timeMillis := utils.UnixMs(fetchResult.Comments[len(fetchResult.Comments)-1].UpdatedAt.Time)
		fetchResult.Next = &timeMillis
		for i := range fetchResult.Comments {
			comment := &fetchResult.Comments[i]
			// keep content hidden if post is hidden
			if comment.Hidden {
				comment.Content = "[removed]"
			}
			// check if user is owner
			if comment.UserID == token.UID {
				comment.Owner = true
			}
		}
	}

	return &fetchResult, nil
}

func (h *handler) handleGetComments(c *gin.Context) {
	// extract request
	var req validation.SaveContentCursor
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	results, err := h.getComments(c, token, req)
	if err != nil {
		// all returned errors are just general client-facing "server errors"
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		return
	}

	// if all goes well, send 200
	response.New(http.StatusOK).Val(results).Send(c)
}
