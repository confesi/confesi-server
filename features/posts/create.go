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
	"gorm.io/gorm"
)

var (
	errorInvalidCategory = errors.New("invalid category")
)

func (h *handler) createPost(c *gin.Context, title string, body string, token *auth.Token, category string) error {
	// start a transaction
	tx := h.db.Begin()
	// if something goes ary, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}()

	// check if category is valid
	var postCategory db.PostCategory
	err := tx.Select("id").Where("name ILIKE ?", category).First(&postCategory).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return serverError
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return errorInvalidCategory
	}

	// fetch the user's facultyId, and schoolId
	var userData db.User
	err = tx.Select("faculty_id, school_id, year_of_study_id").Where("id = ?", token.UID).First(&userData).Error
	if err != nil {
		tx.Rollback()
		return serverError
	}

	// post to save to postgres
	post := db.Post{
		UserID:        token.UID,
		SchoolID:      userData.SchoolID,
		CategoryID:    uint(postCategory.ID.Val),
		FacultyID:     userData.FacultyID,
		YearOfStudyID: userData.YearOfStudyID,
		Title:         title,
		Content:       body,
		Sentiment:     nil,
		Downvote:      0,
		Upvote:        0,
		TrendingScore: 0,
		Hidden:        false,
		// `HottestOn` not included so that it defaults to NULL
	}

	// sentiment analysis of post
	sentiment := AnalyzeText(title + "\n" + body)
	sentimentValue := sentiment.Compound
	if sentimentValue == 0 {
		sentimentValue = sentiment.Neutral
	}
	post.Sentiment = &sentimentValue

	// save user to postgres
	err = tx.Create(&post).Error
	if err != nil {
		tx.Rollback()
		return errors.New(serverError.Error())
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

	err = h.createPost(c, title, body, token, req.Category)
	if err != nil {
		response.New(http.StatusBadRequest).Err(err.Error()).Send(c)
		return
	}

	// if all goes well, send 201
	response.New(http.StatusCreated).Send(c)
}
