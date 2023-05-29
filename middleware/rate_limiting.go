package middleware

import (
	"confesi/config"
	"confesi/lib/response"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// todo: remove so many prints

// Rate limit middleware.
//
// Limits the amount of times a user can access a resource in a given time window.
// Returns a 429 error if the user has exceeded the limit.
func RateLimit(c *gin.Context, tokensPerUnit int, unit time.Duration) {

	var (
		mutex      sync.Mutex
		refillFreq = unit // refill tokens every unit, aka, you have n tokens per unit to use
	)

	var requestStore = config.StoreRef()

	clientIP := c.ClientIP()

	mutex.Lock()
	bucket, ok := requestStore.Bucket[clientIP]
	if !ok {
		fmt.Println("not ok")
		bucket = &config.TokenBucket{
			Tokens:     tokensPerUnit,
			LastRefill: time.Now(),
		}
		requestStore.Bucket[clientIP] = bucket
	} else {
		mutex.Unlock() // Unlock immediately if the IP entry exists
	}

	// refill the tokens for a time interval if a new time window has started
	if !ok && time.Since(bucket.LastRefill) >= refillFreq {
		fmt.Println("refilling bucket")
		mutex.Lock()
		if time.Since(bucket.LastRefill) >= refillFreq { // double-check after acquiring the lock (should be more efficient?)
			bucket.Tokens = tokensPerUnit
			bucket.LastRefill = time.Now()
		}
		mutex.Unlock()
	}

	if bucket.Tokens >= 1 {
		fmt.Println("removing 1 token from bucket")
		bucket.Tokens--
		c.Next()
	} else {
		fmt.Println("too many req!")
		response.New(http.StatusTooManyRequests).
			Err("too many requests").
			Send(c)
	}

	// clean up expired IP entries from the store (efficient?)
	// todo: only cleaning up ip of user on request, not all ips
	if time.Since(bucket.LastRefill) > refillFreq {
		fmt.Println("clean up store")
		mutex.Lock()
		delete(requestStore.Bucket, clientIP)
		mutex.Unlock()
	}
}
