package user

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleClearFaculty(c *gin.Context) {

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	// delete the column "faculty_id" (set to NULL)
	res := h.db.
		Model(&db.User{}).
		Where("id = ?", token.UID).
		Update("faculty_id", nil)
	if res.Error != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}
	if res.RowsAffected == 0 {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
	}

	// say 200 if all goes well
	response.New(http.StatusOK).Send(c)
}
