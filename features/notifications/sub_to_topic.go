package notifications

import (
	"confesi/db"
	"confesi/lib/fire"
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

	// fetch all tokens
	tokens := []db.FcmToken{}
	err = h.db.
		Find(&tokens, "user_id = ?", token.UID).
		Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// todo: validate topic is valid

	// for each token, sub
	for _, t := range tokens {
		fire.SubToTopics(c, h.fb.MsgClient, t.Token, []string{req.Topic})
	}
}
