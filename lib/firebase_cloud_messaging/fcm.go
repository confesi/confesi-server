package fcm

import (
	"confesi/lib/logger"
	"context"
	"errors"
	"fmt"
	"time"

	"firebase.google.com/go/messaging"
	"gorm.io/gorm"
)

var (
	InvalidFcmTokenError      = errors.New("invalid fcm token")
	NoReceivers               = errors.New("no receivers")
	CantSendToTokensAndTopics = errors.New("can't send to both tokens and topics")
	InvalidPayload            = errors.New("invalid payload")
)

const (
	SyncTypeNotificationPrefs = "notification_prefs"
)

type Sender struct {
	Client         *messaging.Client
	Tokens         []string
	Topic          string
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

func (s *Sender) ToTokens(tokens []string) *Sender {
	s.Tokens = tokens
	return s
}

func (s *Sender) ToTopic(topic string) *Sender {
	s.Topic = topic
	return s
}

func (s *Sender) WithData(data map[string]string) *Sender {
	s.Data = data
	return s
}

func (s *Sender) WithMsg(notification *messaging.Notification) *Sender {
	s.Notification = notification
	return s
}

func (s *Sender) Send(db gorm.DB) (error, uint) {
	messages := make([]*messaging.Message, 0)

	if len(s.Tokens) > 0 && s.Topic == "" {
		// Create messages for individual tokens
		for _, token := range s.Tokens {
			message := &messaging.Message{
				Token:        token,
				Data:         s.Data,
				Notification: s.Notification,
				Android: &messaging.AndroidConfig{
					Priority: "high", // default to high to get that sweet "ding"
				},
				APNS:    &messaging.APNSConfig{},
				Webpush: &messaging.WebpushConfig{},
			}
			messages = append(messages, message)
		}
	} else if len(s.Tokens) == 0 && s.Topic != "" {
		// Create a message for the topic
		message := &messaging.Message{
			Topic:        s.Topic,
			Data:         s.Data,
			Notification: s.Notification,
			Android: &messaging.AndroidConfig{
				Priority: "high", // default to high to get that sweet "ding"
			},
			APNS:    &messaging.APNSConfig{},
			Webpush: &messaging.WebpushConfig{},
		}
		messages = append(messages, message)
	} else {
		// Return error if both tokens and topic are present or if both are absent
		return CantSendToTokensAndTopics, 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.ContextTimeout)
	defer cancel()

	if len(messages) == 0 {
		return NoReceivers, 0
	}

	// we need to send in at max bunches of 500, since that's the limit of the FCM API
	tokenBatches := make([][]*messaging.Message, 0)
	if len(messages) > 500 {
		for i := 0; i < len(messages); i += 500 {
			end := i + 500
			if end > len(messages) {
				end = len(messages)
			}
			tokenBatches = append(tokenBatches, messages[i:end])
		}
	} else {
		tokenBatches = append(tokenBatches, messages)
	}

	sends := uint(0)
	deadTokens := make([]string, 0)
	// send each batch
	for _, batch := range tokenBatches {
		batchResponse, _ := s.Client.SendAll(ctx, batch)
		// check the results for each message in the batch
		for j, result := range batchResponse.Responses {
			fmt.Println(result)
			if result.Error != nil {
				if messaging.IsRegistrationTokenNotRegistered(result.Error) {
					deadTokens = append(deadTokens, batch[j].Token)
				}
				sends++
			}
		}
		sends += uint(batchResponse.SuccessCount)
	}

	result := db.Table("fcm_tokens").Where("token IN ?", deadTokens).Delete(nil)
	if result.Error != nil {
		// Handle the error if the deletion fails
		return result.Error, sends
	}

	return nil, sends
}

// "Throw it at the mailbox and hope it gets delivered" approach; no error handling
func SendSyncNotification(db gorm.DB, client *messaging.Client, user_id string, syncType string) {

	var tokens []string

	// Fetch user's tokens from the database
	err := db.Table("users").
		Select("fcm_tokens.token").
		Joins("JOIN fcm_tokens ON fcm_tokens.user_id = users.id").
		Where("users.id = ?", user_id).
		Pluck("fcm_tokens.token", &tokens).
		Error

	if err != nil {
		// handle the error if fetching tokens fails
		logger.StdInfo(fmt.Sprintf("failed to send sync request for set topic prefs: %v", err))
	}

	New(client).ToTokens(tokens).WithData(map[string]string{
		"type": syncType,
	}).Send(db)
}
