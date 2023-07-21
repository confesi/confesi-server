package user

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleSetFaculty(c *gin.Context) {

	// validate the json body from request
	var req validation.UpdateFaculty
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

	// check if user's faculty is valid (aka, the faculty exists in the database)
	faculty := db.Faculty{}
	err = tx.Select("id").Where("faculty = ?", req.Faculty).First(&faculty).Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusBadRequest).Err("faculty doesn't exist").Send(c)
		return
	}
	facultyID := uint8(faculty.ID)

	// update the user's year of study
	res := tx.
		Model(&db.User{}).
		Where("id = ?", token.UID).
		Update("faculty_id", facultyID)
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

	// say 200 if all goes well
	response.New(http.StatusOK).Send(c)
}
