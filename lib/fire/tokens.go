package fire

import (
	"context"
	"fmt"

	"firebase.google.com/go/messaging"
)

func IsValidFcmToken(client *messaging.Client, token string) bool {
	message := &messaging.Message{
		Token: token,
	}

	response, err := client.SendDryRun(context.Background(), message)
	if err != nil {
		// Handle error
		return false
	}

	fmt.Println(response)

	return true
}
