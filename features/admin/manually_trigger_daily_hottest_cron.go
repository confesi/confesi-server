package admin

import (
	"confesi/lib/cronJobs/dailyHottestPosts"
	"confesi/lib/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// TriggerHottestCron godoc
//
//	@Summary		Trigger Daily Hottest Cron
//	@Description	Attempts to execute the cron job once, for the specified date, without retries.
//	@Tags			Admin
//	@Accept			application/json
//	@Produce		application/json
//	@Security		BearerAuth
//	@Security		X-AppCheck-Token
//
//	@Param			day	query		string					true	"Example: 2023-07-09"
//
//	@Success		200	{object}	docs.Success			"Cron Initiated"
//	@Failure		400	{object}	docs.InvalidDateFormat	"Post was Not Found"
//
//	@Router			/admin/daily-hottest-cron [post]
//
// Attempts to execute the cron job once, for the specified date, without retries.
func (h *handler) handleManuallyTriggerDailyHottestCron(c *gin.Context) {
	dateStr := c.Query("day")

	// Parse the date string into a time.Time value
	date, err := time.Parse("2006-01-02", dateStr) // this basically says YYYY-MM-DD, not sure why, but it only works with a dummy date example?
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid date format").Send(c)
		return
	}
	err = dailyHottestPosts.DoDailyHottestJob(date)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		return
	}
	response.New(http.StatusOK).Send(c)
}
