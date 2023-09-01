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
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
)

func (h *handler) handleAddChat(c *gin.Context) {
	// Validate the JSON body from request
	var req validation.AddChat
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	// Get user token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("Failed to get user token").Send(c)
		return
	}

	// Fetch the room to check whether the user is part of the room
	roomSnapshot, err := h.fb.FirestoreClient.Collection("rooms").Doc(req.RoomID).Get(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("Failed to get room data").Send(c)
		return
	}
	var room db.Room
	if err := roomSnapshot.DataTo(&room); err != nil {
		response.New(http.StatusInternalServerError).Err("Error decoding room data").Send(c)
		return
	}

	var userNum int
	var otherUser string
	if room.U1 == token.UID {
		userNum = 1
		otherUser = room.U2
	} else if room.U2 == token.UID {
		userNum = 2
		otherUser = room.U1
	} else {
		response.New(http.StatusBadRequest).Err("User is not part of the room").Send(c)
		return
	}

	chat := db.Chat{
		Msg:    req.Msg,
		RoomID: req.RoomID,
		User:   userNum, // Using "User" instead of "UserID"
		Date:   time.Now().UTC(),
	}

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

		// Update last_msg field in the room document
		updateData := []firestore.Update{
			{Path: "last_msg", Value: chat.Date},
		}
		return tx.Update(roomRef, updateData)
	})

	// Check transaction result
	if err != nil {
		response.New(http.StatusInternalServerError).Err("Failed to complete transaction").Send(c)
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
		response.New(http.StatusInternalServerError).Err("Server error while obtaining FCM tokens").Send(c)
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
