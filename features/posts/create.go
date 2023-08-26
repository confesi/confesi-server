package posts

import (
	"confesi/config"
	"confesi/db"
	"confesi/lib/emojis"
	"confesi/lib/response"
	"confesi/lib/uploads"
	"confesi/lib/utils"
	"errors"
	"fmt"
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

	form, err := c.MultipartForm()
	if err != nil {
		response.New(http.StatusBadRequest).Err("ill-formatted form").Send(c)
		return
	}

	fmt.Println("got here 1")

	fmt.Println("form: ", form)

	files := form.File["files"] // Adjusting this to "files" for multiple uploads
	titles := form.Value["title"]
	bodies := form.Value["body"]
	categories := form.Value["category"]

	fmt.Println("got here 2")

	var title, body, category string
	if len(titles) > 0 {
		title = titles[0]
	}
	if len(bodies) > 0 {
		body = bodies[0]
	}
	if len(categories) > 0 {
		category = categories[0]
	}

	// strip whitespace from title & body
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)

	// input validation & sanitization
	if len(title) == 0 && len(body) == 0 {
		response.New(http.StatusBadRequest).Err("title and body cannot be empty").Send(c)
		return
	}
	if len(category) == 0 || len(category) > 100 { // arbitrary max length to ensure no INSANE value is inputted
		response.New(http.StatusBadRequest).Err("invalid category").Send(c)
		return
	}

	if len(files) > 5 {
		response.New(http.StatusBadRequest).Err("cannot upload more than 5 images").Send(c)
		return
	}
	if len(title) == 0 && len(body) == 0 {
		response.New(http.StatusBadRequest).Err("title and body cannot be empty").Send(c)
		return
	}
	if len(title) > config.TitleMaxLength {
		response.New(http.StatusBadRequest).Err("title too long").Send(c)
		return
	}
	if len(body) > config.BodyMaxLength {
		response.New(http.StatusBadRequest).Err("body too long").Send(c)
		return
	}

	fmt.Println("got here 3")

	imgUrls := []string{}

	fmt.Println("prefix", c.Request.Header.Get("Content-Type")) // todo: temp
	fmt.Println(strings.Contains(c.Request.Header.Get("Content-Type"), "multipart/form-data"))
	fmt.Println("files: ", files)

	// Check if the request's content type is multipart/form-data before trying to read the image
	if strings.Contains(c.Request.Header.Get("Content-Type"), "multipart/form-data") {
		fmt.Println("got here 4")
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			fmt.Println("got here 5")

			// If an error other than http.ErrMissingFile occurs, send an error response
			if err != nil {
				response.New(http.StatusBadRequest).Err("Error reading file").Send(c)
				return
			}

			// Attempt to upload
			fileURL, err := uploads.Upload(file, fileHeader.Filename)
			if err != nil {
				response.New(http.StatusBadRequest).Err(err.Error()).Send(c)
				return
			}

			imgUrls = append(imgUrls, fileURL)

			// Remember to close the file after processing
			file.Close()
		}
	}

	fmt.Println("got here 6")

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
	err = tx.Select("id").Where("name ILIKE ?", category).First(&postCategory).Error
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
		ImgUrl:        &imgUrls[0], // todo: change this to a slice of strings
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
