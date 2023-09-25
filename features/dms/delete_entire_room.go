package dms

// todo: read through, cuz, uh... AI

import (
	"confesi/lib/response"
	"confesi/lib/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

func (h *handler) handleDeleteEntireRoom(c *gin.Context) {
	// Authenticate the user and obtain their token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("failed to get user token").Send(c)
		return
	}

	// Extract roomID from the request
	roomID := c.Param("room-id")
	if roomID == "" {
		response.New(http.StatusBadRequest).Err("room-id parameter required").Send(c)
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

	// Delete all chat messages associated with the roomID
	chatsCollectionRef := h.fb.FirestoreClient.Collection("chats")
	chatsQuery := chatsCollectionRef.Where("room_id", "==", roomID)
	chatsIterator := chatsQuery.Documents(c)

	for {
		chatSnapshot, err := chatsIterator.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			response.New(http.StatusInternalServerError).Err("error fetching chat messages").Send(c)
			return
		}
		_, err = chatsCollectionRef.Doc(chatSnapshot.Ref.ID).Delete(c)
		if err != nil {
			response.New(http.StatusInternalServerError).Err("failed to delete chat message").Send(c)
			return
		}
	}

	// Delete the chat rooms themselves
	roomsIterator := h.fb.FirestoreClient.Collection("rooms").Where("room_id", "==", roomID).Documents(c)
	for {
		roomSnapshot, err := roomsIterator.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			response.New(http.StatusInternalServerError).Err("failed to fetch rooms").Send(c)
			return
		}
		_, err = h.fb.FirestoreClient.Collection("rooms").Doc(roomSnapshot.Ref.ID).Delete(c)
		if err != nil {
			response.New(http.StatusInternalServerError).Err("failed to delete chat room").Send(c)
			return
		}
	}

	// Respond with success
	response.New(http.StatusOK).Send(c)
}
