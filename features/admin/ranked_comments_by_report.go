package admin

import (
	"confesi/config"
	"confesi/lib/logger"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

const (
	seenCommentsCacheExpiry = 24 * time.Hour // one day
)

func (h *handler) handleGetRankedCommentsByReport(c *gin.Context) {
	// extract request
	var req validation.RankedCommentsByReportsQuery
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// session key that can only be created by *this* user, so it can't be guessed to manipulate others' feeds
	commentSpecificKey, err := utils.CreateCacheKey(config.RedisCommentsCacheByReports, token.UID, req.SessionKey)
	if err != nil {
		response.New(http.StatusBadRequest).Err(utils.UuidError.Error()).Send(c)
		return
	}

	if req.PurgeCache {
		// purge the cache
		err := h.redis.Del(c, commentSpecificKey).Err()
		if err != nil {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	// retrieve the seen root comment IDs from the cache
	ids, err := h.redis.SMembers(c, commentSpecificKey).Result()
	if err != nil {
		if err == redis.Nil {
			ids = []string{} // assigns an empty slice
		} else {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	// fetch comments

	// set the expiration for the cache
	err = h.redis.Expire(c, commentSpecificKey, seenCommentsCacheExpiry).Err()
	if err != nil {
		logger.StdErr(err)
		response.New(http.StatusInternalServerError).Err("failed to set cache expiration").Send(c)
		return
	}

	// Send response
	response.New(http.StatusOK).Val(comments).Send(c)
}
