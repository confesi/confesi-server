package hideLog

import (
	"confesi/config"
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

type logDetail struct {
	db.HideLog  `gorm:"embedded"`
	ContentType string `json:"content_type" gorm:"-"`
}

type fetchedReports struct {
	Logs []logDetail `json:"logs"`
	Next *int64      `json:"next"`
}

// References the hide log of the posts/comments you've had hidden by admins. You can view the content even after
// it's been deleted through here if you are the original creator of it.
func (h *handler) handleYourHideLog(c *gin.Context) {
	// extract request
	var req validation.HideLogCursor
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	fetchResults := fetchedReports{}

	err = h.db.
		Where(req.Next.Cursor("created_at <")).
		Where("user_id = ?", token.UID). // just a precaution; the entries here should always be by the creator, but to ensure to not leak any data
		Order("created_at DESC").
		Find(&fetchResults.Logs).
		Limit(config.ReportsPageSize).
		Error

	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	if len(fetchResults.Logs) > 0 {

		timeMicros := (fetchResults.Logs[len(fetchResults.Logs)-1].CreatedAt.Time).UnixMicro()
		fetchResults.Next = &timeMicros
	}

	for i := 0; i < len(fetchResults.Logs); i++ {
		if fetchResults.Logs[i].HideLog.PostID != nil {
			fetchResults.Logs[i].ContentType = "post"
		} else if fetchResults.Logs[i].HideLog.CommentID != nil {
			fetchResults.Logs[i].ContentType = "comment"
		}
	}
	response.New(http.StatusOK).Val(fetchResults).Send(c)
}
