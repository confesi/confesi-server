package admin

import (
	"confesi/config"
	"confesi/db"
	"confesi/lib/response"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

type FetchedDailyHottestCrons struct {
	DailyHottestCrons []db.DailyHottestCron `json:"crons"`
	Next              *uint                 `json:"next"`
}

func (h *handler) handleGetDailyHottestCrons(c *gin.Context) {
	next := c.Query("next")
	var datetime datatypes.Date
	dbQuery := h.db.
		Table("daily_hottest_cron_jobs")

	nextInt, err := strconv.ParseInt(next, 10, 64)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("error parsing next curser").Send(c)
		return
	}

	datetime = datatypes.Date(time.Unix(0, nextInt*int64(time.Millisecond)))
	dbQuery = dbQuery.Where("daily_hottest_cron_jobs.successfully_ran <= ?", datetime)

	fetchResult := FetchedDailyHottestCrons{}

	err = dbQuery.Order("daily_hottest_cron_jobs.successfully_ran DESC").
		Limit(config.DailyHottestCronJobResultsPageSize).Find(&fetchResult.DailyHottestCrons).Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	if len(fetchResult.DailyHottestCrons) > 0 && len(fetchResult.DailyHottestCrons) == config.DailyHottestCronJobResultsPageSize {
		// retrieve the last item's timestamp for the next query
		date := fetchResult.DailyHottestCrons[len(fetchResult.DailyHottestCrons)-1].SuccessfullyRan
		timeValue := time.Time(date)
		// calculate the milliseconds since Unix epoch
		milliseconds := timeValue.UnixNano() / int64(time.Millisecond)
		// assign the milliseconds value to fetchResult.Next
		t := uint(milliseconds - 86400000) // minus a day for the curser
		fetchResult.Next = &t
	}

	response.New(http.StatusOK).Val(fetchResult).Send(c)
}
