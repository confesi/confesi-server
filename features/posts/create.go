package posts

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/validation"
	"fmt"
	"net/http"
	"strings"

	"database/sql"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

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

	// start a transaction
	tx := h.db.Begin()
	// if something went ary, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err("server error").Send(c)
			return
		}
	}()

	// fetch the user's facultyId
	var userData db.User
	err := tx.Select("faculty_id, school_id").Where("id = ?", token.UID).First(&userData).Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
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
		HottestOn:     sql.NullTime{},
	}

	// save user to postgres
	err = tx.Create(&post).Error
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	// commit the transaction
	tx.Commit()

	// if all goes well, send 201
	response.New(http.StatusCreated).Send(c)
}
