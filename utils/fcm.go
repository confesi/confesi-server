package utils

import (
	"context"
	"errors"

	"firebase.google.com/go/messaging"
)

// Default handler to send messages via FCM.
//
// This will help abstract the FCM client and make it easier to send messages with rigit `data` structures in the future.
func SendFcmMsg(client *messaging.Client, token string, topic string, data map[string]string, notification *messaging.Notification) error {
	message := &messaging.Message{
		Data:         data,
		Notification: notification,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{},
	}

	if token != "" {
		message.Token = token
	}

	if topic != "" {
		message.Topic = topic
	}

	if token == "" && topic == "" {
		return errors.New("either token or topic must be provided")
	}

	_, err := client.Send(context.Background(), message)
	if err != nil {
		return err
	}

	return nil
}
