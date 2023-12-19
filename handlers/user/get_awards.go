package user

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleGetAwards(c *gin.Context) {

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	results := []db.AwardsTotal{}
	query := h.db.
		Preload("AwardType").
		Model(db.AwardsTotal{}).
		Where("user_id = ?", token.UID).
		Find(&results).
		Error
	if query != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	response.New(http.StatusOK).Val(results).Send(c)
}
