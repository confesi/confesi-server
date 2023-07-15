package middleware

import (
	"confesi/config"
	"confesi/lib/response"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

const (
	// 30 requests per minute
	tokensPerUnit = 30
	unit          = time.Minute
)

// Rate limit middleware.
//
// Limits the amount of times a user can access a resource in a given time window.
// Returns a 429 error if the user has exceeded the limit.
//
// Includes headers to let the user know how many requests they have left and when the next refill is. Unit: seconds.
func RateLimit(c *gin.Context) {

	store := config.StoreRef()
	clientIP := c.ClientIP()
	ctx := c.Request.Context()

	counter, err := store.Get(ctx, clientIP).Int64()

	// Check whether a cache exists or not
	if err == redis.Nil {
		//If no cache exists create one
		err = store.Set(ctx, clientIP, 1, unit).Err()
		if err != nil {
			response.New(http.StatusInternalServerError).Send(c)
			return
		}
		counter = 1
	} else if err != nil {
		response.New(http.StatusInternalServerError).Send(c)
		return
	}

	// set headers to let the user know know metadata about their rate limit
	c.Header("X-RateLimit-Limit", fmt.Sprint(tokensPerUnit))
	c.Header("X-RateLimit-Remaining", fmt.Sprint(tokensPerUnit-counter))

	// time until next refill
	ttlResult := store.TTL(ctx, clientIP)
	if ttlResult.Err() != nil {
		response.New(http.StatusInternalServerError).Send(c)
		return
	}

	// Retrieve the time left from the result
	ttl, err := ttlResult.Result()
	if err != nil {
		response.New(http.StatusInternalServerError).Send(c)
		return
	}

	c.Header("X-RateLimit-Reset", fmt.Sprint(ttl.String())) // seconds until next refill

	// Determine whether or not user has exceeded the limit
	if counter < tokensPerUnit {
		store.Incr(ctx, clientIP).Result()
		c.Next()
	} else {
		response.New(http.StatusTooManyRequests).
			Err("too many requests").
			Send(c)
	}
}
