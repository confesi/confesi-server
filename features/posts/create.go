package posts

import (
	"confesi/db"
	"confesi/lib/emojis"
	"confesi/lib/response"
	"confesi/lib/uploads"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	invalidCategory = errors.New("invalid category")
)

func (h *handler) handleCreate(c *gin.Context) {

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	// extract request
	var req validation.CreatePostDetails
	err = utils.New(c).ForceCustomTag("required_without", validation.RequiredWithout).Validate(&req)
	if err != nil {
		return
	}

	// Read the image from the request
	imageFile, header, err := c.Request.FormFile("image") // assuming "image" is the field name
	if err != nil && err != http.ErrMissingFile {
		response.New(http.StatusBadRequest).Err("Error reading image").Send(c)
		return
	}

	// If the image exists, attempt to upload
	if imageFile != nil {
		imageURL, err := uploads.Upload(imageFile, header.Filename)
		if err != nil {
			response.New(http.StatusBadRequest).Err(err.Error()).Send(c)
			return
		}
		post.ImageURL = imageURL
	}

	// strip whitespace from title and body (custom validator already confirmed this is still not empty)
	title := strings.TrimSpace(req.Title)
	body := strings.TrimSpace(req.Body)

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
	err = tx.Select("id").Where("name ILIKE ?", req.Category).First(&postCategory).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		response.New(http.StatusBadRequest).Err(invalidCategory.Error()).Send(c)
		return
	}

	// fetch the user's facultyId, and schoolId
	var userData db.User
	err = tx.Select("faculty_id, school_id, year_of_study_id").Where("id = ?", token.UID).First(&userData).Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// post to save to postgres
	post := db.Post{
		UserID:        token.UID,
		SchoolID:      userData.SchoolID,
		CategoryID:    postCategory.ID,
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
	err = tx.Create(&post).Preload("School").Preload("YearOfStudy").Preload("Category").Preload("Faculty").Find(&post).Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// commit the transaction
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// if all goes well, send 201
	response.New(http.StatusCreated).Val(PostDetail{Post: post, UserVote: 0, Owner: true, Emojis: emojis.GetEmojis(&post)}).Send(c)
}
