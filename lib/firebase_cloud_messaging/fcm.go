package fcm

import (
	"confesi/db"
	"confesi/lib/logger"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"firebase.google.com/go/messaging"
)

var (
	InvalidFcmTokenError      = errors.New("invalid fcm token")
	NoReceivers               = errors.New("no receivers")
	CantSendToTokensAndTopics = errors.New("can't send to both tokens and topics")
	InvalidPayload            = errors.New("invalid payload")
)

type Sender struct {
	Client           *messaging.Client
	Tokens           []string
	Topic            string
	Data             map[string]string
	Notification     *messaging.Notification
	ContextTimeout   time.Duration
	ContentAvailable bool
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

// Sets if it should be sent as a background message only.
//
// Defaults to true.
func (s *Sender) ShownInBackgroundOnly(onlyBackground bool) *Sender {
	s.ContentAvailable = !onlyBackground
	return s
}

func (s *Sender) Send() (error, uint) {
	messages := make([]*messaging.Message, 0)

	apnsConfig := &messaging.APNSConfig{
		Headers: map[string]string{
			"method": "POST",
			"apns-priority": func() string {
				if s.ContentAvailable {
					return "10" // Higher priority for immediate display
				}
				return "5" // Lower priority for background processing
			}(),
			"apns-push-type": func() string {
				if s.ContentAvailable {
					return "alert" // Show on screen with sound
				}
				return "background" // Background processing without alert
			}(),
			"apns-collapse-id": "confesi",
			"apns-expiration":  "0",
		},
		Payload: &messaging.APNSPayload{
			Aps: &messaging.Aps{
				Sound: func() string {
					if s.ContentAvailable {
						return "default" // Play sound for alerts
					}
					return "default" // No sound for background processing
				}(),
				ContentAvailable: s.ContentAvailable,
				Alert: func() *messaging.ApsAlert {
					if s.ContentAvailable {
						return &messaging.ApsAlert{
							Title: s.Notification.Title,
							Body:  s.Notification.Body,
						}
					}
					return nil
				}(),
			},
		},
	}

	androidConfig := &messaging.AndroidConfig{
		Notification: &messaging.AndroidNotification{
			ChannelID: "confesi",
			Sound:     "default",
		},
		Priority: func() string {
			if s.ContentAvailable {
				return "high" // High priority for alerts
			}
			return "normal" // Normal priority for background processing
		}(),
	}

	if len(s.Tokens) > 0 && s.Topic == "" {
		// Create messages for individual tokens
		for _, token := range s.Tokens {
			apnsConfig.Headers["path"] = "/3/device/" + token
			message := &messaging.Message{
				FCMOptions: &messaging.FCMOptions{
					AnalyticsLabel: "confesi",
				},
				Token:        token,
				Data:         s.Data,
				Notification: s.Notification,
				Android:      androidConfig,
				APNS:         apnsConfig,
			}

			test, err := s.Client.Send(context.Background(), message)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(test)

			messages = append(messages, message)
		}
	} else if len(s.Tokens) == 0 && s.Topic != "" {
		// Create a message for the topic
		message := &messaging.Message{
			Topic:        s.Topic,
			Data:         s.Data,
			Notification: s.Notification,
			Android:      androidConfig,
			APNS:         apnsConfig,
			Webpush:      &messaging.WebpushConfig{},
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
			if result.Error != nil {
				if messaging.IsRegistrationTokenNotRegistered(result.Error) {
					deadTokens = append(deadTokens, batch[j].Token)
				}
			}
		}
		sends += uint(batchResponse.SuccessCount)
	}

	if len(deadTokens) > 0 {
		result := db.New().Table("fcm_tokens").Where("token IN ?", deadTokens).Delete(&db.FcmToken{})
		if result.Error != nil {
			// Handle the error if the deletion fails
			return result.Error, sends
		}
	}

	data, err := json.Marshal(s.Data)
	if err != nil {
		return InvalidPayload, sends
	}

	dataString := string(data)

	var userIDs []string

	db.New().Table("fcm_tokens").Where("token IN ?", s.Tokens).Pluck("user_id", &userIDs)

	var logs []db.NotificationLog

	for _, user := range userIDs {
		logs = append(logs, db.NotificationLog{
			UserID: user,
			Body:   s.Notification.Body,
			Title:  s.Notification.Title,
			Data:   dataString,
		})
	}

	// Create Log
	db.New().Table("notification_logs").Create(&logs)

	// log how many sends
	logger.StdInfo(fmt.Sprintf("sent %d of %d fcm messages successfully", sends, len(messages)))

	return nil, sends
}
