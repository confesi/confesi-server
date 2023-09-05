package dms

import (
	"confesi/lib/response"
	"confesi/lib/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

func (h *handler) handleDeleteChat(c *gin.Context) {
	// Authenticate the user and obtain their token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("failed to get user token").Send(c)
		return
	}

	// Extract chatID from the request
	chatID := c.Param("chat-id")
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
	roomQuery := h.fb.FirestoreClient.Collection("rooms").
		Where("room_id", "==", roomID).
		Where("user_id", "==", token.UID)

	roomSnapshot, err := roomQuery.Documents(c).Next()
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

	// Respond with success
	response.New(http.StatusOK).Send(c)
}
