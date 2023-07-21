package middleware

import (
	"confesi/lib/response"
	"confesi/lib/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// Rate limit middleware based on the user's UID.
//
// ! Assumes the "user" will be set into context already!!
func UidRateLimit(c *gin.Context, tokensPerUnit int64, unit time.Duration, rootKey string) {

	store := StoreRef()
	ctx := c.Request.Context()

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	idSessionKey := rootKey + ":" + token.UID

	counter, err := store.Get(ctx, idSessionKey).Int64()

	// Check whether a cache exists or not
	if err == redis.Nil {
		//If no cache exists create one
		err = store.Set(ctx, idSessionKey, 0, unit).Err()
		if err != nil {
			response.New(http.StatusInternalServerError).Send(c)
			return
		}
		counter = 1
	} else if err != nil {
		response.New(http.StatusInternalServerError).Send(c)
		return
	}

	// Determine whether or not user has exceeded the limit
	if counter < tokensPerUnit {
		store.Incr(ctx, idSessionKey).Result()
		c.Next()
	} else {
		response.New(http.StatusTooManyRequests).
			Err("too many routed requests").
			Send(c)
	}
}