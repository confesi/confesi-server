package middleware

import (
	"confesi/lib/response"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type routedRequestRateLimit struct {
	Limit             string `json:"limit"`
	RemainingRequests string `json:"remaining_requests"`
	ResetInSeconds    string `json:"reset_in_seconds"`
}

// Rate limit middleware based on the user's UID.
//
// ! Assumes the "user" will be set into context already!!
func RoutedRateLimit(c *gin.Context, tokensPerUnit int64, unit time.Duration, rootKey string, identifier string, tooManyRequestsErrMsg string) {

	store := StoreRef()
	ctx := c.Request.Context()

	idSessionKey := rootKey + ":" + identifier

	counter, err := store.Get(ctx, idSessionKey).Int64()

	// check whether a cache exists or not
	if err == redis.Nil {
		// if no cache exists create one
		err = store.Set(ctx, idSessionKey, 0, time.Minute*30).Err()
		if err != nil {
			response.New(http.StatusInternalServerError).Send(c)
			return
		}
		counter = 1
	} else if err != nil {
		response.New(http.StatusInternalServerError).Send(c)
		return
	}

	// time until next refill
	ttlResult := store.TTL(ctx, idSessionKey)
	if ttlResult.Err() != nil {
		response.New(http.StatusInternalServerError).Send(c)
		return
	}

	// retrieve the time left from the result
	ttl, err := ttlResult.Result()
	if err != nil {
		response.New(http.StatusInternalServerError).Send(c)
		return
	}

	// Determine whether or not user has exceeded the limit
	if counter < tokensPerUnit {
		store.Incr(ctx, idSessionKey).Result()
		c.Next()
	} else {
		response.New(http.StatusTooManyRequests).
			Err(tooManyRequestsErrMsg).
			Val(routedRequestRateLimit{
				Limit:             fmt.Sprint(tokensPerUnit),
				RemainingRequests: fmt.Sprint(tokensPerUnit - counter),
				ResetInSeconds:    fmt.Sprintf("%.0f", ttl.Seconds()),
			}).
			Send(c)
	}
}
