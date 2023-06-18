package utils

const (
	Comments = "comments"
	Posts    = "posts"
)

func CreateCacheKey(contentType string, userID string, sessionID string) string {
	return contentType + ":" + userID + "[" + sessionID + "]"
}
