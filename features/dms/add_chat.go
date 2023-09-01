package dms

import (
	"confesi/config/builders"
	"confesi/db"
	fcm "confesi/lib/firebase_cloud_messaging"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"context"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
)

func (h *handler) handleAddChat(c *gin.Context) {
	// msg
	// room_id
	// ensuring that the user is part of the room

	// Validate the JSON body from request
	var req validation.AddChat
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	// Get user token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// If token.UID is not a subset of room_id, return error
	// This works because the room_id is a combination of the two user ids with the post ID messaged over
	if !strings.Contains(req.RoomID, token.UID) {
		response.New(http.StatusBadRequest).Err("user not part of the room").Send(c)
		return
	}

	chat := db.Chat{
		Msg:    req.Msg,
		RoomID: req.RoomID,
		UserID: token.UID,
		Date:   time.Now().UTC(),
	}

	var otherUser string

	// Define the transaction function
	err = h.fb.FirestoreClient.RunTransaction(c, func(ctx context.Context, tx *firestore.Transaction) error {
		// Get reference to chats and rooms collections
		chatsCollectionRef := h.fb.FirestoreClient.Collection("chats")
		roomRef := h.fb.FirestoreClient.Collection("rooms").Doc(req.RoomID)

		// Add the chat to Firestore
		_, _, err := chatsCollectionRef.Add(c, chat)
		if err != nil {
			return err
		}

		// Fetch the other user from the room document
		doc, err := tx.Get(roomRef)
		if err != nil {
			return err
		}
		if val, exists := doc.Data()["UserOther"]; exists && val != token.UID {
			otherUser = val.(string)
		}

		// Update last_msg field in the room document
		updateData := []firestore.Update{
			{Path: "last_msg", Value: chat.Date},
		}
		return tx.Update(roomRef, updateData)
	})

	// Check transaction result
	if err != nil {
		// Handle the Firestore error here
		response.New(http.StatusInternalServerError).Err("failed to complete transaction").Send(c)
		return
	}

	// Obtain FCM tokens for the affected other user
	var tokens []string
	err = h.db.
		Table("fcm_tokens").
		Select("fcm_tokens.token").
		Joins("JOIN users ON users.id = fcm_tokens.user_id").
		Where("users.id = ?", otherUser).
		Pluck("fcm_tokens.token", &tokens).
		Error

	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	go fcm.New(h.fb.MsgClient).
		ToTokens(tokens).
		WithMsg(builders.AdminSendNotificationNoti("title", "body")).
		WithData(map[string]string{}).
		Send()

	// Send a success response
	response.New(http.StatusOK).Send(c)
}
