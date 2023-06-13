package saves

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/validation"
	"fmt"
	"net/http"
	"time"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const (
	CursorSize = 10 // how many posts/comments to fetch at a time
)

type fetchSavedResult struct {
	Posts []db.Post `json:"content"`
	Next  *int64    `json:"next"`
}

func (h *handler) fetchSavedContent(c *gin.Context, token *auth.Token, req validation.SaveContentCursor) (fetchSavedResult, error) {
	next := time.UnixMilli(int64(req.Next))
	// todo: convert to work with both posts & comments
	var posts []db.Post
	err := h.db.
		Joins("JOIN saved_posts ON saved_posts.post_id = posts.id").
		Where("saved_posts.updated_at < ?", next).
		Where("saved_posts.user_id = ?", token.UID).
		Order("saved_posts.updated_at DESC").
		Limit(CursorSize).
		Find(&posts).Error
	if err != nil {
		return fetchSavedResult{}, err
	}

	// determine the next value for the cursor
	var nextCursor *int64
	if len(posts) > 0 {
		lastPost := posts[len(posts)-1]
		nextCursorValue := lastPost.UpdatedAt.UnixMilli()
		nextCursor = &nextCursorValue
	}

	result := fetchSavedResult{
		Posts: posts,
		Next:  nextCursor,
	}

	return result, nil
}

func (h *handler) handleGetSaved(c *gin.Context) {
	// extract request
	var req validation.SaveContentCursor

	// create validator
	validator := validator.New()

	// create a binding instance with the validator, check if json valid, if so, deserialize into req
	binding := &validation.DefaultBinding{
		Validator: validator,
	}
	if err := binding.Bind(c.Request, &req); err != nil {
		response.New(http.StatusBadRequest).Err(fmt.Sprintf("failed validation: %v", err)).Send(c)
		return
	}

	// TODO: START: once firebase user utils function is merged, use that instead to be cleaner
	// get firebase user
	user, ok := c.Get("user")
	if !ok {
		response.New(http.StatusInternalServerError).Err(ServerError).Send(c)
		return
	}

	token, ok := user.(*auth.Token)
	if !ok {
		response.New(http.StatusInternalServerError).Err(ServerError).Send(c)
		return
	}
	// TODO: END

	results, err := h.fetchSavedContent(c, token, req)
	if err != nil {
		// all returned errors are just general client-facing "server errors"
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		return
	}

	// if all goes well, send 200
	response.New(http.StatusOK).Val(results).Send(c)
}
