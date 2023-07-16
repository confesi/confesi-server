package config

import (
	"confesi/lib/cache"
	"time"

	"github.com/go-redis/redis/v8"
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
