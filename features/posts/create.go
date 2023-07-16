package posts

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
	"net/http"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

//	CreatePost godoc
//
//	@Summary		Create a Post.
//	@Description	Create a post with the given title and body.
//	@Tags			Posts
//	@Accept			application/json
//	@Produce		application/json
//
// @Security BearerAuth
//
//	@Param			Body	body		string	true	"The title and body of the post"	SchemaExample({\n "title": "My Title",\n "body": "My Body"\n})
//
//	@Success		200		{string}	string	"ok"
//	@Failure		500		{object}	string	"server error"
//
//	@Router			/posts/create [post]
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
		// `HottestOn` not included so that it defaults to NULL
	}

	// save user to postgres
	err = tx.Create(&post).Error
	if err != nil {
		tx.Rollback()
		return errors.New("server error")
	}

	// commit the transaction
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return serverError
	}
	return nil
}

func (h *handler) handleCreate(c *gin.Context) {

	// extract request
	var req validation.CreatePostDetails
	err := utils.New(c).ForceCustomTag("required_without", validation.RequiredWithout).Validate(&req)
	if err != nil {
		return
	}

	// strip whitespace from title and body (custom validator already confirmed this is still not empty)
	title := strings.TrimSpace(req.Title)
	body := strings.TrimSpace(req.Body)

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	err = h.createPost(c, title, body, token)
	if err != nil {
		// all returned errors are just general client-facing "server errors"
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	// if all goes well, send 201
	response.New(http.StatusCreated).Send(c)
}
