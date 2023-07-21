package schools

import (
	"confesi/db"
	"confesi/lib/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleGetRandomSchool(c *gin.Context) {

	school := db.School{}

	err := h.DB.Order("RANDOM()").First(&school).Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	response.New(http.StatusOK).Val(school).Send(c)
}
