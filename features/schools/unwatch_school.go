package schools

import (
	"confesi/db"
	"confesi/lib/encryption"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

func (h *handler) unwatchSchool(c *gin.Context, token *auth.Token, req validation.WatchSchool, unmaskedId uint) error {
	school := db.SchoolFollow{
		UserID:   token.UID,
		SchoolID: db.EncryptedID{Val: unmaskedId},
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

	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	unmaskedId, err := encryption.Unmask(req.SchoolID)
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid id").Send(c)
		return
	}

	err = h.unwatchSchool(c, token, req, unmaskedId)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
	}
	response.New(http.StatusOK).Send(c)
}
