package posts

import (
	"confesi/config"
	"confesi/db"
	"confesi/lib/response"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	serverError = errors.New("server error")
)

func (h *handler) getHottestPosts(c *gin.Context, date time.Time) ([]db.Post, error) {
	var posts []db.Post
	err := h.db.
		Where("hottest_on = ?", date).
		Limit(config.HottestPostsSize).
		Preload("School").
		Preload("Faculty").
		Order("vote_score DESC"). // fetches the hottest X posts for the day, and comparatively between them, ranks them by `vote_score`
		Find(&posts).
		Error
	if err != nil {
		return nil, serverError
	}
	return posts, nil
}

func (h *handler) handleGetHottest(c *gin.Context) {
	dateStr := c.Query("day")

	// Parse the date string into a time.Time value
	date, err := time.Parse("2006-01-02", dateStr) // this basically says YYYY-MM-DD, not sure why, but it only works with a dummy date example?
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid date format").Send(c)
		return
	}

	posts, err := h.getHottestPosts(c, date)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		return
	}
	response.New(http.StatusOK).Val(posts).Send(c)
}
