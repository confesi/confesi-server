package saves

import (
	"confesi/config"
	"confesi/handlers/comments"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

type FetchedComments struct {
	Comments []comments.CommentDetail `json:"comments"`
	Next     *int64                   `json:"next"`
}

func (h *handler) getComments(c *gin.Context, token *auth.Token, req validation.SaveContentCursor) (*FetchedComments, error) {
	fetchResult := FetchedComments{}

	query := `
    SELECT comments.*,
        COALESCE(
            (
                SELECT votes.vote
                FROM votes
                WHERE votes.comment_id = comments.id
                    AND votes.user_id = ?
                LIMIT 1
            ),
            '0'::vote_score_value
        ) AS user_vote,
        TRUE as saved,  -- because we're joining with saved_comments
        EXISTS(
            SELECT 1
            FROM reports
            WHERE reports.comment_id = comments.id
            AND reports.reported_by = ?
        ) as reported
    FROM comments
    JOIN saved_comments ON comments.id = saved_comments.comment_id
    WHERE saved_comments.user_id = ?
        ` + req.Next.Cursor("AND saved_comments.created_at <") + `
        AND comments.hidden = false
    ORDER BY saved_comments.created_at DESC
    LIMIT ?
    `

	err := h.db.Raw(query, token.UID, token.UID, token.UID, config.SavedPostsAndCommentsPageSize).
		Find(&fetchResult.Comments).Error

	if err != nil {
		return nil, err
	}

	if len(fetchResult.Comments) > 0 {
		timeMicros := (fetchResult.Comments[len(fetchResult.Comments)-1].Comment.CreatedAt.Time).UnixMicro()
		fetchResult.Next = &timeMicros
		for i := range fetchResult.Comments {
			comment := &fetchResult.Comments[i]
			// check if user is owner
			if comment.Comment.UserID == token.UID {
				comment.Owner = true
			}
			if !utils.ProfanityEnabled(c) {
				comment.Comment = comment.Comment.CensorComment()
			}
			comment.Comment.ObscureIfHidden()
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
