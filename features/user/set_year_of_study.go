package user

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *handler) handleSetYearOfStudy(c *gin.Context) {

	// validate the json body from request
	var req validation.UpdateYearOfStudy
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	// start a transaction
	tx := h.db.Begin()
	// if something goes ary, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err("server error").Send(c)
			return
		}
	}()

	// check if user's year of study is valid (aka, the year of study exists in the database)
	yearOfStudy := db.YearOfStudy{}
	err = tx.Raw("SELECT id FROM year_of_study WHERE name ILIKE ?", req.YearOfStudy).First(&yearOfStudy).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		response.New(http.StatusBadRequest).Err("invalid year of study").Send(c)
		return
	}
	yearOfStudyID := uint8(yearOfStudy.ID)

	// update the user's year of study
	res := tx.
		Model(&db.User{}).
		Where("id = ?", token.UID).
		Update("year_of_study_id", yearOfStudyID)
	if res.Error != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}
	if res.RowsAffected == 0 {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	// commit the transaction
	err = tx.Commit().Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	// say 200 if all goes well
	response.New(http.StatusOK).Send(c)
}
