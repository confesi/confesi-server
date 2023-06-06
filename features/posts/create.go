package posts

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/validation"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"database/sql"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func (h *handler) createPost(c *gin.Context, title string, body string, token *auth.Token) error {
	// start a transaction
	tx := h.db.Begin()
	// if something goes ary, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err("server error").Send(c)
			return
		}
	}()

	// fetch the user's facultyId, and schoolId
	var userData db.User
	err := tx.Select("faculty_id, school_id").Where("id = ?", token.UID).First(&userData).Error
	if err != nil {
		tx.Rollback()
		return errors.New("server error")
	}

	// post to save to postgres
	post := db.Post{
		UserID:        token.UID,
		SchoolID:      userData.SchoolID,
		FacultyID:     userData.FacultyID,
		Title:         title,
		Content:       body,
		Downvote:      0,
		Upvote:        0,
		TrendingScore: 0,
		Hidden:        false,
		HottestOn:     sql.NullTime{}, // default to a null time, aka, it hasn't yet been hottest on any day
	}

	// save user to postgres
	err = tx.Create(&post).Error
	if err != nil {
		tx.Rollback()
		return errors.New("server error")
	}

	// commit the transaction
	tx.Commit()
	return nil
}

func (h *handler) handleCreate(c *gin.Context) {

	// extract request
	var req validation.CreatePostDetails

	// create validator
	validator := validator.New()

	// register custom tag to ensure that at minimum either the title or body is present
	validator.RegisterValidation("required_without", validation.RequiredWithout)

	// create a binding instance with the validator, check if json valid, if so, deserialize into req
	binding := &validation.DefaultBinding{
		Validator: validator,
	}
	if err := binding.Bind(c.Request, &req); err != nil {
		response.New(http.StatusBadRequest).Err(fmt.Sprintf("failed validation: %v", err)).Send(c)
		return
	}

	// strip whitespace from title and body (custom validator already confirmed this is still not empty)
	title := strings.TrimSpace(req.Title)
	body := strings.TrimSpace(req.Body)

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

	err := h.createPost(c, title, body, token)
	if err != nil {
		// all returned errors are just general client-facing "server errors"
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	// if all goes well, send 201
	response.New(http.StatusCreated).Send(c)
}
