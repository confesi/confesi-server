package saves

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/validation"
	"errors"
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

type FetchResult interface {
	AddPosts(posts []db.Post, next int64)
	AddComments(comments []db.Comment, next int64)
	GetSavedTable() string
	GetSavedIdName() string
	GetReferencedTable() string
}

type FetchedPosts struct {
	Posts []db.Post `json:"posts"`
	Next  int64     `json:"next"`
}

func (f *FetchedPosts) AddPosts(posts []db.Post, next int64) {
	// loop through all posts and add, and for the last added post, update the Next field
	f.Posts = posts
	if len(posts) > 0 {
		f.Next = posts[len(posts)-1].CreatedAt.UnixMilli()
	} else {
		f.Next = next
	}
}

func (f *FetchedPosts) AddComments(comments []db.Comment, next int64) {} // Ignoring, just to implement interface

func (f *FetchedPosts) GetSavedTable() string {
	return "saved_posts"
}

func (f *FetchedPosts) GetReferencedTable() string {
	return "posts"
}

func (f *FetchedPosts) GetSavedIdName() string {
	return "post_id"
}

type FetchedComments struct {
	Comments []db.Comment `json:"comments"`
	Next     int64        `json:"next"`
}

func (f *FetchedComments) isFetchResult() {}

func (f *FetchedComments) AddPosts(posts []db.Post, next int64) {} // Ignoring, just to implement interface

func (f *FetchedComments) AddComments(comments []db.Comment, next int64) {
	f.Comments = comments
	if len(comments) > 0 {
		f.Next = comments[len(comments)-1].CreatedAt.UnixMilli()
	} else {
		f.Next = next
	}
}

func (f *FetchedComments) GetReferencedTable() string {
	return "comments"
}

func (f *FetchedComments) GetSavedTable() string {
	return "saved_comments"
}

func (f *FetchedComments) GetSavedIdName() string {
	return "comment_id"
}

// todo: add a new the migration (so far I directly modified migration 1)
func (h *handler) fetchSavedContent(c *gin.Context, token *auth.Token, req validation.SaveContentCursor) (FetchResult, error) {
	var posts []db.Post
	var comments []db.Comment
	var fetchResult FetchResult
	var model interface{}
	if req.ContentType == "post" {
		fetchResult = &FetchedPosts{
			Posts: []db.Post{},
		}
		model = &posts
	} else if req.ContentType == "comment" {
		fetchResult = &FetchedComments{
			Comments: []db.Comment{},
		}
		model = &comments
	} else {
		return nil, errors.New(ServerError)
	}
	next := time.UnixMilli(int64(req.Next))

	savedTable := fetchResult.GetSavedTable()
	savedTableContentId := fetchResult.GetSavedIdName()
	referencedTable := fetchResult.GetReferencedTable()

	err := h.db.
		Joins("JOIN "+savedTable+" ON "+referencedTable+".id = "+savedTable+"."+savedTableContentId).
		Table(referencedTable).
		Where(savedTable+".updated_at < ?", next).
		Where(savedTable+".user_id = ?", token.UID).
		Where(referencedTable+".hidden = ?", false).
		Order(savedTable + ".updated_at DESC").
		Limit(CursorSize).
		Find(model).Error

	if err != nil {
		return fetchResult, nil
	}

	// only one of these will ever do anything
	fetchResult.AddPosts(posts, int64(req.Next))
	fetchResult.AddComments(comments, int64(req.Next))

	return fetchResult, nil
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
