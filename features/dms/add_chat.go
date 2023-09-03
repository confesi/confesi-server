package dms

import (
	"confesi/config/builders"
	"confesi/db"
	fcm "confesi/lib/firebase_cloud_messaging"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"context"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

func (h *handler) handleAddChat(c *gin.Context) {
	// Get user token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("failed to get user token").Send(c)
		return
	}

	// extract request
	var req validation.AddChat
	err = utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	// Identify the other user's room document
	otherUserRoomQuery := h.fb.FirestoreClient.Collection("rooms").
		Where("room_id", "==", req.RoomID).
		Where("user_id", "!=", token.UID)

	otherUserRoomSnapshotIterator := otherUserRoomQuery.Documents(c)
	otherUserRoomSnapshot, err := otherUserRoomSnapshotIterator.Next()

	if err == iterator.Done {
		response.New(http.StatusBadRequest).Err("other user's room not found with given criteria").Send(c)
		return
	} else if err != nil {
		fmt.Println(err)
		response.New(http.StatusInternalServerError).Err("error querying other user's room").Send(c)
		return
	}

	var otherUserRoom db.Room
	if err := otherUserRoomSnapshot.DataTo(&otherUserRoom); err != nil {
		response.New(http.StatusInternalServerError).Err("failed decoding other user's room data").Send(c)
		return
	}

	var chat db.Chat
	chat.RoomID = otherUserRoom.RoomID
	chat.UserNumber = otherUserRoom.UserNumber
	chat.Date = time.Now().UTC()
	chat.Msg = req.Msg

	// Define the transaction function
	err = h.fb.FirestoreClient.RunTransaction(c, func(ctx context.Context, tx *firestore.Transaction) error {
		// Get reference to chats collection
		chatsCollectionRef := h.fb.FirestoreClient.Collection("chats")

		// Directly get the roomRef from the already retrieved otherUserRoomSnapshot
		roomRef := otherUserRoomSnapshot.Ref

		// Add the chat to Firestore
		_, _, err = chatsCollectionRef.Add(c, chat)
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
		fmt.Println(err)
		response.New(http.StatusInternalServerError).Err("failed to complete transaction").Send(c)
		return
	}

	// Obtain FCM tokens for the affected user (other user in this case)
	var tokens []string
	err = h.db.
		Table("fcm_tokens").
		Select("fcm_tokens.token").
		Joins("JOIN users ON users.id = fcm_tokens.user_id").
		Where("users.id = ?", otherUserRoom.UserID).
		Pluck("fcm_tokens.token", &tokens).
		Error

	if err != nil {
		response.New(http.StatusInternalServerError).Err("failed to get FCM tokens").Send(c)
		return
	}

	go fcm.New(h.fb.MsgClient).
		ToTokens(tokens).
		WithMsg(builders.NewChatNoti(req.Msg, otherUserRoom.Name)).
		WithData(builders.NewChatData(chat.RoomID)).
		Send()

	// Send a success response
	response.New(http.StatusOK).Send(c)
}
