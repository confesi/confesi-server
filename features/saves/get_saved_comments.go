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

type FetchedComments struct {
	Comments []db.Comment `json:"comments"`
	Next     *int64       `json:"next"`
}

func (h *handler) getComments(c *gin.Context, token *auth.Token, req validation.SaveContentCursor) (*FetchedComments, error) {
	fetchResult := FetchedComments{}
	next := time.UnixMilli(int64(req.Next))

	err := h.db.
		Joins("JOIN saved_comments ON comments.id = saved_comments.comment_id").
		Table("comments").
		Select("comments.*, saved_comments.updated_at").
		Where("saved_comments.updated_at < ?", next).
		Where("saved_comments.user_id = ?", token.UID).
		Where("comments.hidden = ?", false).
		Order("saved_comments.updated_at DESC").
		Limit(cursorSize).
		Find(&fetchResult.Comments).Error

	if err != nil {
		return nil, err
	}

	if len(fetchResult.Comments) > 0 {
		timeMillis := fetchResult.Comments[len(fetchResult.Comments)-1].UpdatedAt.UnixMilli()
		fetchResult.Next = &timeMillis
	}

	return &fetchResult, nil
}

func (h *handler) handleGetComments(c *gin.Context) {
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

	results, err := h.getComments(c, token, req)
	if err != nil {
		// all returned errors are just general client-facing "server errors"
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		return
	}

	// if all goes well, send 200
	response.New(http.StatusOK).Val(results).Send(c)
}
