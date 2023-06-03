package posts

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/validation"
	"fmt"
	"net/http"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func (h *handler) handleCreate(c *gin.Context) {

	// extract request
	var req validation.CreatePostDetails

	// create a binding instance with the validator, check if json valid, if so, deserialize into req
	binding := &validation.DefaultBinding{
		Validator: validator.New(),
	}
	if err := binding.Bind(c.Request, &req); err != nil {
		response.New(http.StatusBadRequest).Err(fmt.Sprintf("failed validation: %v", err)).Send(c)
		return
	}

	// strip whitespace from title and body
	title := strings.TrimSpace(req.Title)
	body := strings.TrimSpace(req.Body)

	// check if title and body are BOTH empty
	if title == "" && body == "" {
		response.New(http.StatusBadRequest).Err("title and body cannot both be empty").Send(c)
		return
	}

	user, ok := c.Get("user")
	if !ok {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	token, ok := user.(*auth.Token)
	if !ok {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	// fetch the user's facultyId
	var userFaculty db.User
	err := h.db.Select("faculty_id").Where("id = ?", token.UID).First(&userFaculty).Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err("error fetching user's faculty").Send(c)
		return
	}

	// post to save to postgres
	post := db.Post{
		UserID:        token.UID,
		FacultyID:     userFaculty.FacultyID,
		Title:         title,
		Content:       body,
		Downvote:      0,
		Upvote:        0,
		TrendingScore: 0,
	}

	// save user to postgres
	err = h.db.Create(&post).Error
	if err != nil {
		response.New(http.StatusBadRequest).Err("error creating post").Send(c)
		return
	}

	// if all goes well, send 201
	response.New(http.StatusCreated).Send(c)
}
