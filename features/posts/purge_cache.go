package posts

import (
	"confesi/config"
	"confesi/lib/response"
	"confesi/lib/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PurgePostCache godoc
//
//	@Summary		Purge Cache.
//	@Description	Deletes the cache for the posts of a user based on their session.
//	@Tags			Posts
//	@Accept			application/json
//	@Produce		application/json
//	@Security		BearerAuth
//	@Security		X-AppCheck-Token
//
//	@Param			session-key	query		string				true	"Example: 6ba7b810-9dad-11d1-80b4-00c04fd430c8"
//
//	@Success		200			{object}	docs.CachePurged	"Cache Purged"
//	@Failure		500			{object}	docs.ServerError	"Server Error"
//
//	@Router			/posts/purge [delete]
//
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

	cacheKey, err := utils.CreateCacheKey(config.RedisPostsCache, token.UID, sessionKey)
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
