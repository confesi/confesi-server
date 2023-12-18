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

type reportDetail struct {
	db.Report   `gorm:"embedded"`
	ContentType string `json:"content_type" gorm:"-"`
}

type fetchResults struct {
	Reports []reportDetail `json:"reports"`
	Next    *int64         `json:"next"`
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
		timeMicros := (fetchResults.Reports[len(fetchResults.Reports)-1].Report.CreatedAt.Time).UnixMicro()
		fetchResults.Next = &timeMicros
	}

	for i := 0; i < len(fetchResults.Reports); i++ {
		if fetchResults.Reports[i].Report.PostID != nil {
			fetchResults.Reports[i].ContentType = "post"
		} else if fetchResults.Reports[i].Report.CommentID != nil {
			fetchResults.Reports[i].ContentType = "comment"
		}
	}

	response.New(http.StatusOK).Val(fetchResults).Send(c)
}
