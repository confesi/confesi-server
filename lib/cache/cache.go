package cache

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
)

var redisClient *redis.Client

func init() {
	connString := os.Getenv("REDIS_CONN")
	if connString == "" {
		panic("`REDIS_CONN` env not set")
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr: connString,
	})
	if _, err := redisClient.Ping(context.TODO()).Result(); err != nil {
		panic(fmt.Sprintf("failed to connect to redis: %v", err))
	}
}

func New() *redis.Client {
	return redisClient
}
