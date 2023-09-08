package notifications

import (
	"confesi/config"
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleGetNotifications(c *gin.Context) {

	var req validation.YourNotificationsQuery
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	var fetchResults []db.NotificationLog

	err = h.db.
		Where("user_id = ?", token.UID).
		Where(req.Next.Cursor("updated_at >")).
		Order("updated_at ASC").
		Find(&fetchResults).
		Limit(config.ViewYourNotificationsPageSize).
		Error

	response.New(http.StatusOK).Val(fetchResults).Send(c)
}
