package user

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleSetSchool(c *gin.Context) {

	// validate the json body from request
	var req validation.UpdateSchool
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

	// check if user's school is valid (aka, the school exists in the database)
	school := db.School{}
	err = tx.Select("id").Where("name ILIKE ?", req.FullSchoolName).First(&school).Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusBadRequest).Err("school doesn't exist").Send(c)
		return
	}

	schoolID := uint8(school.ID)

	// update the user's school
	res := tx.
		Model(&db.User{}).
		Where("id = ?", token.UID).
		Update("school_id", schoolID)
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
