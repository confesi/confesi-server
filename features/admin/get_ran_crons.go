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

// GetRanCrons godoc
//
//	@Summary		Get Ran Crons.
//	@Description	Obtain the Ran Cron Jobs of a specific type
//	@Tags			Admin
//	@Accept			application/json
//	@Produce		application/json
//	@Security		BearerAuth
//	@Security		X-AppCheck-Token
//	@Param			Body	body		string				true	"Pagination Cursor and Type"	SchemaExample({\n    "next": 1689156632542916,\n    "type": "all"\n})
//	@Success		200		{object}	docs.GetCronJobs	"Cron Jobs Ran Result"
//	@Failure		500		{object}	docs.ServerError	"Server Error"
//
//	@Router			/admin/crons [get]
func (h *handler) handleGetDailyHottestCrons(c *gin.Context) {
	// extract request
	var req validation.FetchRanCrons
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	fetchResults := FetchedCronJobs{}

	query := h.db.
		Where(req.Next.Cursor("created_at <"))

	if req.Type != "all" {
		query = query.Where("type = ?", req.Type)
	}

	err = query.
		Order("created_at DESC").
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
