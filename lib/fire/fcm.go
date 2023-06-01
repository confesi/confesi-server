package fire

import (
	"context"
	"errors"
	"time"

	"firebase.google.com/go/messaging"
)

type FcmMsgSender struct {
	Client         *messaging.Client
	Token          string
	Topic          string
	Data           map[string]string
	Notification   *messaging.Notification
	ContextTimeout time.Duration
}

func NewSender(client *messaging.Client) *FcmMsgSender {
	return &FcmMsgSender{
		Client:         client,
		ContextTimeout: 5 * time.Second,
	}
}

func (s *FcmMsgSender) WithToken(token string) *FcmMsgSender {
	s.Token = token
	return s
}

func (s *FcmMsgSender) WithTopic(topic string) *FcmMsgSender {
	s.Topic = topic
	return s
}

func (s *FcmMsgSender) WithData(data map[string]string) *FcmMsgSender {
	s.Data = data
	return s
}

func (s *FcmMsgSender) WithNotification(notification *messaging.Notification) *FcmMsgSender {
	s.Notification = notification
	return s
}

func (s *FcmMsgSender) Send() error {
	message := &messaging.Message{
		Data:         s.Data,
		Notification: s.Notification,
		Android: &messaging.AndroidConfig{
			Priority: "high", // default to high to get that sweet "ding"
		},
		APNS: &messaging.APNSConfig{}, // TODO: set APNS config here once we get the Apple Dev Account (tracked in Issues)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.ContextTimeout)
	defer cancel()

	if s.Token != "" {
		message.Token = s.Token
	}

	if s.Topic != "" {
		message.Topic = s.Topic
	}

	if s.Token == "" && s.Topic == "" {
		return errors.New("either token or topic must be provided")
	}

	_, err := s.Client.Send(ctx, message)
	if err != nil {
		return err
	}

	return nil
}
