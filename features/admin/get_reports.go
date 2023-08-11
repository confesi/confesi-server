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

type FetchedReports struct {
	Reports []db.Report `json:"reports"`
	Next    *int64      `json:"next"`
}

func (h *handler) handleGetReports(c *gin.Context) {
	// extract request
	var req validation.FetchReports
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	userRoles, err := getUserRoles(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		return
	}

	fetchResults := FetchedReports{}

	query := h.db.
		Where(req.Next.Cursor("created_at <"))

	if len(userRoles.SchoolMods) > 0 {
		query.Where("school_id IN ?", userRoles.SchoolMods)
	}

	if req.Type != "all" {
		query = query.Where("type = ?", req.Type)
	}

	err = query.
		Preload("ReportType").
		Joins("JOIN report_types ON report_types.id = reports.type_id").
		Order("created_at DESC").
		Find(&fetchResults.Reports).
		Limit(config.ReportsPageSize).
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
