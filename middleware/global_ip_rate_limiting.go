package middleware

import (
	"confesi/config"
	"confesi/lib/cache"
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

var store *redis.Client

func init() {
	store = cache.New() // Redis client
}

type Bucket struct {
	Tokens         int
	LastRefill     time.Time
	RefillInterval time.Duration
}

func StoreRef() *redis.Client {
	return store
}

// Rate limit middleware.
//
// Limits the amount of times a user can access a resource in a given time window.
// Returns a 429 error if the user has exceeded the limit.
//
// Includes headers to let the user know how many requests they have left and when the next refill is. Unit: seconds.
func RateLimit(c *gin.Context) {

	store := StoreRef()
	clientIP := c.ClientIP()
	ctx := c.Request.Context()

	idSessionKey := config.RedisRateLimitingCache + ":" + clientIP

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

	// set headers to let the user know know metadata about their rate limit
	c.Header("X-RateLimit-Limit", fmt.Sprint(tokensPerUnit))
	c.Header("X-RateLimit-Remaining", fmt.Sprint(tokensPerUnit-counter))

	// time until next refill
	ttlResult := store.TTL(ctx, idSessionKey)
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

	c.Header("X-RateLimit-Reset", fmt.Sprint(ttl.Seconds())) // seconds until next refill

	// Determine whether or not user has exceeded the limit
	if counter < tokensPerUnit {
		store.Incr(ctx, idSessionKey).Result()
		c.Next()
	} else {
		response.New(http.StatusTooManyRequests).
			Err("too many ip requests").
			Send(c)
	}
}
