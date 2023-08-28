package posts

import (
	"confesi/config"
	tags "confesi/lib/emojis"
	"confesi/lib/response"
	"confesi/lib/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type DailyHottestReturn struct {
	Posts []PostDetail `json:"posts"`
	Date  string       `json:"date"`
}

func (h *handler) handleGetHottest(c *gin.Context) {

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	dateStr := c.Query("day")

	var date time.Time
	if dateStr == "" {
		// Fetch the most recent day with hottest posts
		err = h.db.Raw("SELECT MAX(hottest_on) FROM posts WHERE hidden = ?", false).Scan(&date).Error
		if err != nil {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
		dateStr = date.Format("2006-01-02")
	} else {
		// Parse the date string into a time.Time value
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			response.New(http.StatusBadRequest).Err("invalid date format").Send(c)
			return
		}
	}

	var posts []PostDetail
	err = h.db.
		Where("hottest_on = ?", date).
		Where("hidden = ?", false).
		Limit(config.HottestPostsPageSize).
		Preload("School").
		Preload("Category").
		Preload("Faculty").
		Preload("YearOfStudy").
		Order("trending_score DESC").
		Select(`
			posts.*,
			COALESCE(
				(
					SELECT votes.vote
					FROM votes
					WHERE votes.post_id = posts.id
					AND votes.user_id = ?
					LIMIT 1
				),
				'0'::vote_score_value
			) AS user_vote
		`, token.UID).
		Find(&posts).
		Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	for i := range posts {
		post := &posts[i]
		// check if the user is the owner of each post
		if post.UserID == token.UID {
			post.Owner = true
		}
		if !utils.ProfanityEnabled(c) {
			post.Post = post.Post.CensorPost()
		}
		post.Emojis = tags.GetEmojis(&post.Post)
	}

	response.New(http.StatusOK).Val(DailyHottestReturn{Posts: posts, Date: dateStr}).Send(c)
}
