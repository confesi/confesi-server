package admin

import (
	"confesi/config"
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FetchedCronJobs struct {
	Crons []db.CronJob `json:"crons"`
	Next  *int64       `json:"next"`
}

func (h *handler) handleGetDailyHottestCrons(c *gin.Context) {
	// extract request
	var req validation.FetchRanCrons
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	fetchResults := FetchedCronJobs{}

	err = h.db.
		Where(req.Next.Cursor("created_at >")).
		Order("created_at ASC").
		Find(&fetchResults.Crons).
		Limit(config.CronJobPageSize).
		Error

	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	if len(fetchResults.Crons) > 0 {
		timeMicros := (fetchResults.Crons[len(fetchResults.Crons)-1].CreatedAt.Time).UnixMicro()
		fetchResults.Next = &timeMicros
	}

	response.New(http.StatusOK).Val(fetchResults).Send(c)
}
