package utils

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// DeleteCacheFolder deletes all keys in the specified 'folder'
func DeleteCacheFolder(c *context.Context, redis *redis.Client, folder string) error {
	// Obtain all keys in the folder
	redis_keys, err := redis.Keys(*c, folder+":*").Result()
	if err != nil {
		return err
	}

	for _, key := range redis_keys {
		err = redis.Del(*c, key).Err() // Delete each key
		if err != nil {
			return err
		}
	}
	return nil
}
