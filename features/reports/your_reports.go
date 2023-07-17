package reports

import (
	"confesi/config"
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

type fetchResults struct {
	Reports []db.Report `json:"reports"`
	Next    *int64      `json:"next"`
}

func (h *handler) handleGetYourReports(c *gin.Context) {
	// extract request
	var req validation.ReportCursor
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	fetchResults := fetchResults{}

	err = h.db.
		Preload("ReportType").
		Where("reported_by = ?", token.UID).
		Where(req.Next.Cursor("created_at >")).
		Order("created_at ASC").
		Find(&fetchResults.Reports).
		Limit(config.ViewYourReportsPageSize).
		Error

	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	if len(fetchResults.Reports) > 0 {
		timeMicros := (fetchResults.Reports[len(fetchResults.Reports)-1].CreatedAt.Time).UnixMicro()
		fetchResults.Next = &timeMicros
	}

	response.New(http.StatusOK).Val(fetchResults).Send(c)
}
