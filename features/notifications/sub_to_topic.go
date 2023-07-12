package notifications

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleSubToTopic(c *gin.Context) {
	// validate request
	var req validation.FcmTopicQuery
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// todo: FK to ensure there's a match between either valid sub type or name of watchd uni ? Or just check if valid topic?

	// fetch all topic
	topic := db.FcmTopic{
		UserID: token.UID,
		Name:   req.Topic,
	}
	err = h.db.
		Where("user_id = ? AND name = ?", token.UID, req.Topic).
		FirstOrCreate(&topic).
		Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	response.New(http.StatusOK).Send(c)
}
