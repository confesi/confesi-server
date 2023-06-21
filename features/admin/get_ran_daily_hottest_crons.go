package admin

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"
	"time"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

const (
	cursorSize = 10
)

// todo: make work

type FetchedDailyHottestCrons struct {
	DailyHottestCrons []db.DailyHottestCron `json:"crons"`
	Next              *datatypes.Date       `json:"next"`
}

func (h *handler) getDailyHottestCrons(c *gin.Context, token *auth.Token, req validation.DailyHottestCronsCursor) (FetchedDailyHottestCrons, error) {
	fetchResult := FetchedDailyHottestCrons{}

	next := req.Next
	datetime := datatypes.Date(time.Unix(0, int64(next)*int64(time.Millisecond)))

	err := h.db.
		Table("daily_hottest_cron_jobs").
		Where("daily_hottest_cron_jobs.successfully_ran < ?", datetime).
		Order("daily_hottest_cron_jobs.successfully_ran DESC").
		Limit(cursorSize).
		Find(&fetchResult.DailyHottestCrons).Error

	if err != nil {
		return fetchResult, err
	}

	if len(fetchResult.DailyHottestCrons) > 0 {
		date := &fetchResult.DailyHottestCrons[len(fetchResult.DailyHottestCrons)-1].SuccessfullyRan
		fetchResult.Next = date
	}

	return fetchResult, nil
}

func (h *handler) handleGetDailyHottestCrons(c *gin.Context) {

	var req validation.DailyHottestCronsCursor
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	results, err := h.getDailyHottestCrons(c, token, req)
	if err != nil {
		// all returned errors are just general client-facing "server errors"
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		return
	}

	// if all goes well, send 200
	response.New(http.StatusOK).Val(results).Send(c)
}
