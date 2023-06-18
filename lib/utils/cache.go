package utils

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

const (
	Comments        = "comments"
	Posts           = "posts"
	CacheExpiration = 24 * time.Hour
)

func ContentTypeToContentKey(contentType string) (string, error) {
	if contentType == Comments || contentType == Posts {
		return contentType, nil
	}
	return "", errors.New("invalid content type")
}

func SaveToCache(c *gin.Context, redisClient *redis.Client, contentType string, sessionID string, content []string) error {
	contentKey, err := ContentTypeToContentKey(contentType)
	if err != nil {
		return err
	}
	err = redisClient.HSet(c, contentKey, sessionID, content).Err()
	if err != nil {
		return err
	}
	return nil
}

func PurgeFromCache(c *gin.Context, redisClient *redis.Client, contentType string, sessionID string) error {
	contentKey, err := ContentTypeToContentKey(contentType)
	if err != nil {
		return err
	}
	// delete the contentKey's sessionID
	err = redisClient.HDel(c, contentKey, sessionID).Err()
	if err != nil {
		return err
	}
	return nil
}

func GetFromCache(c *gin.Context, redisClient *redis.Client, contentType string, sessionID string) (string, error) {
	contentKey, err := ContentTypeToContentKey(contentType)
	if err != nil {
		return "", err
	}
	// get the contentKey's sessionID
	content, err := redisClient.HGet(c, contentKey, sessionID).Result()
	if err != nil {
		return "", err
	}
	return content, nil
}
