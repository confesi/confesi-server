package config

import (
	"sync"
	"time"
)

type requestStore struct {
	Bucket map[string]*TokenBucket
	Mutex  sync.Mutex
}

var store requestStore

func init() {
	store = requestStore{Bucket: make(map[string]*TokenBucket)}
}

type TokenBucket struct {
	Tokens     int
	LastRefill time.Time
}

func StoreRef() *requestStore {
	return &store
}
