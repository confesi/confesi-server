package saves

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"fmt"
	"net/http"
	"time"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type FetchedPosts struct {
	Posts []db.Post `json:"posts"`
	Next  *int64    `json:"next"`
}

func (h *handler) getPosts(c *gin.Context, token *auth.Token, req validation.SaveContentCursor) (FetchedPosts, error) {
	fetchResult := FetchedPosts{}
	next := time.UnixMilli(int64(req.Next))

	err := h.db.
		Joins("JOIN saved_posts ON posts.id = saved_posts.post_id").
		Table("posts").
		Select("posts.*, saved_posts.updated_at").
		Where("saved_posts.updated_at < ?", next).
		Where("saved_posts.user_id = ?", token.UID).
		Where("posts.hidden = ?", false).
		Order("saved_posts.updated_at DESC").
		Preload("School").
		Preload("Faculty").
		Limit(cursorSize).
		Find(&fetchResult.Posts).Error

	if err != nil {
		return fetchResult, err
	}

	if len(fetchResult.Posts) > 0 {
		timeMillis := fetchResult.Posts[len(fetchResult.Posts)-1].UpdatedAt.UnixMilli()
		fetchResult.Next = &timeMillis
	}

	return fetchResult, nil
}

func (h *handler) handleGetPosts(c *gin.Context) {
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
