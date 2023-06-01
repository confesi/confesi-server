package config

import (
	"sync"
	"time"
)

var store sync.Map

func init() {
	store = sync.Map{} // thread-safe sync map for holding rate limiting buckets
}

type Bucket struct {
	Tokens         int
	LastRefill     time.Time
	RefillInterval time.Duration
}

func StoreRef() *sync.Map {
	return &store
}
