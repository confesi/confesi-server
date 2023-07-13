package notifications

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleUnsubToPriv(c *gin.Context) {
	// validate request
	var req validation.FcmPrivQuery
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	var delField string
	if req.ContentType == "post" {
		delField = "post_id"
	} else if req.ContentType == "comment" {
		delField = "comment_id"
	} else {
		// should never happen with validated struct, but to be defensive
		response.New(http.StatusBadRequest).Err(fmt.Sprintf("invalid content type")).Send(c)
		return
	}

	fcmTopic := db.FcmPriv{}

	err = h.db.
		Delete(&fcmTopic, "user_id = ? AND "+delField+" = ?", token.UID, req.ContentID).
		Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	response.New(http.StatusOK).Send(c)
}
