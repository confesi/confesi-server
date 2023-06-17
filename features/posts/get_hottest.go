package posts

import (
	"confesi/db"
	"confesi/lib/response"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	serverError             = errors.New("server error")
	maxHottestPostsReturned = 5 // this should never be more than 5, but to be defensive, we'll set a limit anyway. This is because X posts max/day should ever be set to hottest.
)

func (h *handler) getHottestPosts(c *gin.Context, date time.Time) ([]db.Post, error) {
	var posts []db.Post
	err := h.db.Where("hottest_on = ?", date).Limit(maxHottestPostsReturned).Find(&posts).Error
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
		response.New(http.StatusBadRequest).Err("Invalid date format").Send(c)
		return
	}

	posts, err := h.getHottestPosts(c, date)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		return
	}
	response.New(http.StatusOK).Val(posts).Send(c)
}
