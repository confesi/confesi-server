package fcm

import (
	"context"

	"firebase.google.com/go/messaging"
)

func IsValidFcmToken(client *messaging.Client, token string) bool {
	message := &messaging.Message{
		Token: token,
	}

	_, err := client.SendDryRun(context.Background(), message)
	if err != nil {
		// Handle error
		return false
	}

	return true
}
