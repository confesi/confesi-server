package dms

import (
	"confesi/config/builders"
	fcm "confesi/lib/firebase_cloud_messaging"
	"confesi/lib/response"
	"confesi/lib/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

// todo: put into firestore transaction to ensure atomicy

func (h *handler) handleDeleteChat(c *gin.Context) {
	// Authenticate the user and obtain their token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("failed to get user token").Send(c)
		return
	}

	// Extract chatID from the request
	chatID := c.Query("id")
	if chatID == "" {
		response.New(http.StatusBadRequest).Err("chat-id parameter required").Send(c)
		return
	}

	// Fetch the chat message
	chatRef := h.fb.FirestoreClient.Collection("chats").Doc(chatID)
	chatSnapshot, err := chatRef.Get(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("failed to fetch chat message").Send(c)
		return
	}

	// Extract the roomID from the chat message
	roomID, ok := chatSnapshot.Data()["room_id"].(string)
	if !ok {
		response.New(http.StatusInternalServerError).Err("failed to extract room ID from chat message").Send(c)
		return
	}

	// Ensure the user is part of the specified chat room
	thisUsersRoomQuery := h.fb.FirestoreClient.Collection("rooms").
		Where("room_id", "==", roomID).
		Where("user_id", "==", token.UID)

	// Get the other user's room
	otherUserRoomQuery := h.fb.FirestoreClient.Collection("rooms").
		Where("room_id", "==", roomID).
		Where("user_id", "!=", token.UID)

	// Get the other user's room
	otherUserRoomSnapshot, err := otherUserRoomQuery.Documents(c).Next()
	if err == iterator.Done || otherUserRoomSnapshot == nil {
		response.New(http.StatusBadRequest).Err("user not part of the specified chat room").Send(c)
		return
	} else if err != nil {
		response.New(http.StatusInternalServerError).Err("error checking chat room membership").Send(c)
		return
	}

	otherUserUid := otherUserRoomSnapshot.Data()["user_id"].(string)
	if otherUserUid == "" {
		response.New(http.StatusInternalServerError).Err("failed to extract other user ID from chat message").Send(c)
		return
	}

	roomSnapshot, err := thisUsersRoomQuery.Documents(c).Next()
	if err == iterator.Done || roomSnapshot == nil {
		response.New(http.StatusBadRequest).Err("user not part of the specified chat room").Send(c)
		return
	} else if err != nil {
		response.New(http.StatusInternalServerError).Err("error checking chat room membership").Send(c)
		return
	}

	// Delete the specific chat message
	_, err = chatRef.Delete(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("failed to delete chat message").Send(c)
		return
	}

	// spin up light-weight thread to send FCM message

	var tokens []string
	err = h.db.
		Table("fcm_tokens").
		Select("fcm_tokens.token").
		Joins("JOIN users ON users.id = fcm_tokens.user_id").
		Where("users.id = ?", otherUserUid).
		Pluck("fcm_tokens.token", &tokens).
		Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err("failed to get FCM tokens").Send(c)
		return
	}

	// (don't handle error case since it's not necessary)
	go fcm.New(h.fb.MsgClient).
		ToTokens(tokens).
		WithMsg(builders.DeletedChatNoti()).
		WithData(builders.DeletedChatData()).
		Send()

	// Respond with success
	response.New(http.StatusOK).Send(c)
}
