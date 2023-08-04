package fcm

import (
	"firebase.google.com/go/messaging"
	"github.com/gin-gonic/gin"
)

// Sub a token to topics
//
// Return: list of topics that were failed in attempting to sub to
func SubToTopics(c *gin.Context, client *messaging.Client, fcmToken string, topics []string) []string {
	failedSubs := []string{}
	for i := range topics {
		response, err := client.SubscribeToTopic(c, []string{fcmToken}, topics[i])
		if err != nil || response.SuccessCount != 1 {
			failedSubs = append(failedSubs, topics[i])
		}
	}
	return topics
}

// Unsubs a token from topics
//
// Return: list of topics that failed to unsub
func UnsubToTopics(c *gin.Context, client *messaging.Client, fcmToken string, topics []string) []string {
	failedUnsubs := []string{}
	for i := range topics {
		response, err := client.UnsubscribeFromTopic(c, []string{fcmToken}, topics[i])
		if err != nil || response.SuccessCount != 1 {
			failedUnsubs = append(failedUnsubs, topics[i])
		}
	}
	return topics
}
