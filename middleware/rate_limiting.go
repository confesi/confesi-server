package middleware

import (
	"confesi/lib"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// todo: remove nested functions if possible

// Rate limit middleware.
//
// Limits the amount of times a user can access a resource in a given time window.
// Returns a 429 error if the user has exceeded the limit.
func RateLimit(c *gin.Context, tokensPerUnit int, unit time.Duration) gin.HandlerFunc {

	// todo: lift the tokenBucket and requestStore out of this function into global state,
	// todo: or something similar because it's getting reset everytime right now

	type tokenBucket struct {
		tokens     int
		lastRefill time.Time
	}

	var (
		mutex        sync.Mutex
		requestStore = make(map[string]*tokenBucket) // map of tokens each ip has
		refillFreq   = unit                          // refill tokens every unit, aka, you have n tokens per unit to use
	)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		mutex.Lock()
		bucket, ok := requestStore[clientIP]
		if !ok {
			bucket = &tokenBucket{
				tokens:     tokensPerUnit,
				lastRefill: time.Now(),
			}
			requestStore[clientIP] = bucket
			mutex.Unlock() // Unlock before executing further code
		} else {
			mutex.Unlock() // Unlock immediately if the IP entry exists
		}

		// refill the tokens for a time interval if a new time window has started
		if !ok && time.Since(bucket.lastRefill) >= refillFreq {
			mutex.Lock()
			if time.Since(bucket.lastRefill) >= refillFreq { // double-check after acquiring the lock (should be more efficient?)
				bucket.tokens = tokensPerUnit
				bucket.lastRefill = time.Now()
			}
			mutex.Unlock()
		}

		if bucket.tokens >= 1 {
			bucket.tokens--
			c.Next()
		} else {
			lib.New(http.StatusTooManyRequests).
				Err("too many requests").
				Send(c)
		}

		// clean up expired IP entries from the requestStore (efficient?)
		if time.Since(bucket.lastRefill) > refillFreq {
			mutex.Lock()
			delete(requestStore, clientIP)
			mutex.Unlock()
		}
	}
}
