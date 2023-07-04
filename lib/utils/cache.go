package utils

import (
	"errors"

	"github.com/google/uuid"
)

var (
	UuidError = errors.New("invalid UUID")
)

func CreateCacheKey(contentType string, userID string, sessionID string) (string, error) {
	_, err := uuid.Parse(sessionID) // must be a valid UUID
	if err != nil {
		return "", errors.New("invalid session ID")
	}

	return contentType + ":" + userID + "[" + sessionID + "]", nil
}
