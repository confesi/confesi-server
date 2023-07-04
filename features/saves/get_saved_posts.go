package saves

import (
	"confesi/features/posts"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"
	"time"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

type FetchedPosts struct {
	Posts []posts.PostDetail `json:"posts"`
	Next  *int64             `json:"next"`
}

func (h *handler) getPosts(c *gin.Context, token *auth.Token, req validation.SaveContentCursor) (*FetchedPosts, error) {
	fetchResult := FetchedPosts{}
	next := time.UnixMilli(int64(req.Next))
	query := `
	SELECT posts.*, saved_posts.updated_at,
		COALESCE(
			(
				SELECT votes.vote
				FROM votes
				WHERE votes.post_id = posts.id
					AND votes.user_id = ?
				LIMIT 1
			),
			'0'::vote_score_value
		) AS user_vote
	FROM posts
	JOIN saved_posts ON posts.id = saved_posts.post_id
	WHERE saved_posts.updated_at < ?
		AND saved_posts.user_id = ?
		AND posts.hidden = false
	ORDER BY saved_posts.updated_at DESC
	LIMIT ?
`

	err := h.db.Raw(query, token.UID, next, token.UID, cursorSize).
		Preload("School").
		Preload("Faculty").
		Find(&fetchResult.Posts).Error

	if err != nil {
		return nil, err
	}

	if len(fetchResult.Posts) > 0 {
		timeMillis := fetchResult.Posts[len(fetchResult.Posts)-1].UpdatedAt.UnixMilli()
		fetchResult.Next = &timeMillis
		for i := range fetchResult.Posts {
			post := &fetchResult.Posts[i]
			// keep content hidden if post is hidden
			if post.Hidden {
				post.Content = "[removed]"
			}
			// check if user is owner
			if post.UserID == token.UID {
				post.Owner = true
			}
		}
	}

	return &fetchResult, nil
}

func (h *handler) handleGetPosts(c *gin.Context) {
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

	results, err := h.getPosts(c, token, req)
	if err != nil {
		// all returned errors are just general client-facing "server errors"
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		return
	}

	// if all goes well, send 200
	response.New(http.StatusOK).Val(results).Send(c)
}
