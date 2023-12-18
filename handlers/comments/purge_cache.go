package comments

import (
	"confesi/config"
	"confesi/lib/response"
	"confesi/lib/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Deletes the cache for the comments of a user based on their session.
//
// Useful to prevent storing the comments of a user if they don't need it stored.
//
// Only root comments are stored in cache (dynamic sort criteria), because replies are paginated through by stable `created_at` sort key.
func (h *handler) handlePurgeCommentsCache(c *gin.Context) {
	// get the session key
	sessionKey := c.Query("session-key")

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	cacheKey, err := utils.CreateCacheKey(config.RedisCommentsCache, token.UID, sessionKey)
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
