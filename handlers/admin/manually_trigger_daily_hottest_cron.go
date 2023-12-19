package admin

import (
	"confesi/lib/cronJobs/dailyHottestPosts"
	"confesi/lib/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Attempts to execute the cron job once, for the specified date, without retries.
func (h *handler) handleManuallyTriggerDailyHottestCron(c *gin.Context) {
	dateStr := c.Query("day")

	// Parse the date string into a time.Time value
	date, err := time.Parse("2006-01-02", dateStr) // this basically says YYYY-MM-DD, not sure why, but it only works with a dummy date example?
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid date format").Send(c)
		return
	}
	err = dailyHottestPosts.DoDailyHottestJob(date, h.fb)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		return
	}
	response.New(http.StatusOK).Send(c)
}
