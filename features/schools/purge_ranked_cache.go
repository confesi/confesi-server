package schools

import (
	"confesi/lib/response"
	"confesi/lib/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handlePurgeRankedSchoolsCache(c *gin.Context) {
	// get the session key
	sessionKey := c.Query("session-key")

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	cacheKey, err := utils.CreateCacheKey("schools_rank", token.UID, sessionKey)
	if err != nil {
		response.New(http.StatusBadRequest).Err(utils.UuidError.Error()).Send(c)
		return
	}

	err = h.redis.Del(c, cacheKey).Err()
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	response.New(http.StatusOK).Send(c)
}
