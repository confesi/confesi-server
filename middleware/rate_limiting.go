package middleware

import (
	"confesi/config"
	"confesi/lib/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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
func RateLimit(c *gin.Context) {
	refillFreq := unit // refill tokens every unit, aka, you have n tokens per unit to use

	store := config.StoreRef()

	clientIP := c.ClientIP()

	bucketValue, ok := store.Load(clientIP)
	var bucket *config.Bucket
	if ok {
		bucket = bucketValue.(*config.Bucket)
	} else {
		bucket = &config.Bucket{
			Tokens:         tokensPerUnit,
			LastRefill:     time.Now(),
			RefillInterval: unit,
		}
		store.Store(clientIP, bucket)
	}

	// clean up expired entries if they've been expired for more than 2 times the time unit
	store.Range(func(key, value interface{}) bool {
		ip := key.(string)
		entry := value.(*config.Bucket)
		if time.Since(entry.LastRefill) > 2*unit {
			store.Delete(ip)
		}
		return true
	})

	// refill the tokens for a time interval if a new time window has started
	if time.Since(bucket.LastRefill) >= refillFreq {
		bucket.Tokens = tokensPerUnit
		bucket.LastRefill = time.Now()
	}

	if bucket.Tokens >= 1 {
		bucket.Tokens--
		c.Next()
	} else {
		response.New(http.StatusTooManyRequests).
			Err("too many requests").
			Send(c)
	}
}
