package fcm

// todo: use REGISTRATION TOKENS for specific post subscriptions, and TOPICS for general subscriptions?
// todo: how to handle TOPICS here? I ignored them for now?

import (
	"context"
	"errors"
	"time"

	"firebase.google.com/go/messaging"
)

var (
	InvalidFcmTokenError = errors.New("invalid fcm token")
	InvalidReceivers     = errors.New("invalid receivers")
	InvalidPayload       = errors.New("invalid payload")
)

type Sender struct {
	Client         *messaging.Client
	Tokens         []string
	Topics         []string
	Data           map[string]string
	Notification   *messaging.Notification
	ContextTimeout time.Duration
}

func New(client *messaging.Client) *Sender {
	return &Sender{
		Client:         client,
		ContextTimeout: 5 * time.Second,
	}
}

func (s *Sender) WithToken(tokens []string) *Sender {
	s.Tokens = tokens
	return s
}

func (s *Sender) WithTopic(topics []string) *Sender {
	s.Topics = topics
	return s
}

func (s *Sender) WithData(data map[string]string) *Sender {
	s.Data = data
	return s
}

func (s *Sender) WithNotification(notification *messaging.Notification) *Sender {
	s.Notification = notification
	return s
}

func (s *Sender) Send() error {
	messages := make([]*messaging.Message, 0)
	for _, token := range s.Tokens {
		messages = append(messages, &messaging.Message{
			Token:        token,
			Data:         s.Data,
			Notification: s.Notification,
			Android: &messaging.AndroidConfig{
				Priority: "high", // default to high to get that sweet "ding"
			},
			APNS:    &messaging.APNSConfig{}, // TODO: set APNS config here once we get the Apple Dev Account (tracked in Issues)
			Webpush: &messaging.WebpushConfig{},
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.ContextTimeout)
	defer cancel()

	if len(s.Tokens) == 0 && len(s.Topics) == 0 {
		return InvalidReceivers
	}

	if s.Data == nil && s.Notification == nil {
		return InvalidPayload
	}

	_, err := s.Client.SendAll(ctx, messages)
	if err != nil {
		return err
	}

	return nil
}
