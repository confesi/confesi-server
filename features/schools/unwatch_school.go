package schools

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"fmt"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func (h *handler) unwatchSchool(c *gin.Context, token *auth.Token, req validation.WatchSchool) error {
	school := db.SchoolFollow{
		UserID:   token.UID,
		SchoolID: req.SchoolID,
	}
	err := h.DB.Delete(&school, "user_id = ? AND school_id = ?", school.UserID, school.SchoolID).Error
	if err != nil {
		return serverError
	}
	return nil
}

func (h *handler) handleUnwatchSchool(c *gin.Context) {
	// extract request
	var req validation.WatchSchool

	// create validator
	validator := validator.New()

	binding := &validation.DefaultBinding{
		Validator: validator,
	}
	if err := binding.Bind(c.Request, &req); err != nil {
		response.New(http.StatusBadRequest).Err(fmt.Sprintf("failed validation: %v", err)).Send(c)
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	err = h.unwatchSchool(c, token, req)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
	}
	response.New(http.StatusOK).Send(c)
}
