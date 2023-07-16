package user

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *handler) handleGetUser(c *gin.Context) {

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	user := db.User{}
	err = h.db.
		Preload("School").
		Preload("Faculty").
		Find(&user, "id = ?", token.UID).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		response.New(http.StatusNotFound).Err("user not found").Send(c)
		return
	}
	err, mod := db.ModLevelToString(user.ModID)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	err, year := db.YearOfStudyToString(uint(user.YearOfStudy))
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	user.Mod = mod
	user.StudyYear = year

	response.New(http.StatusOK).Val(user).Send(c)
	return
}
