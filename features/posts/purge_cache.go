package posts

import (
	"confesi/lib/response"
	"confesi/lib/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Deletes the cache for the posts of a user based on their session.
//
// Useful to prevent storing the posts of a user if they don't need it stored.
func (h *handler) handlePurgePostsCache(c *gin.Context) {
	// get the session key
	sessionKey := c.Query("session-key")

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	cacheKey := utils.CreateCacheKey("posts", token.UID, sessionKey)
	err = h.redis.Del(c, cacheKey).Err()
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	response.New(http.StatusOK).Send(c)
}