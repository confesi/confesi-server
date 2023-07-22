package admin

import (
	"confesi/config"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleFetchReportForPostById(c *gin.Context) {
	// extract request
	var req validation.FetchReportsForPostById
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	fetchResults := fetchResults{}

	err = h.db.
		Preload("ReportType").
		Where(req.Next.Cursor("created_at >")).
		Where("post_id IS NOT NULL").
		Order("created_at ASC").
		Find(&fetchResults.Reports).
		Limit(config.AdminViewAllReportsPerPostId).
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