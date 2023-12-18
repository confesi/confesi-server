package dms

import (
	"confesi/lib/response"
	"confesi/lib/utils"
	"context"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

func (h *handler) handleClearEntireChat(c *gin.Context) {
	// Authenticate the user and obtain their token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("failed to get user token").Send(c)
		return
	}

	// Extract roomID from the request
	roomID := c.Query("room-id")
	if roomID == "" {
		response.New(http.StatusBadRequest).Err("room-id query param required").Send(c)
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

	// Delete all chat messages associated with the roomID using a transaction
	err = h.fb.FirestoreClient.RunTransaction(c, func(ctx context.Context, tx *firestore.Transaction) error {
		chatsCollectionRef := h.fb.FirestoreClient.Collection("chats")
		chatsQuery := chatsCollectionRef.Where("room_id", "==", roomID)
		chatsIterator := chatsQuery.Documents(c)

		for {
			chatSnapshot, err := chatsIterator.Next()
			if err == iterator.Done {
				break
			} else if err != nil {
				return err
			}
			err = tx.Delete(chatsCollectionRef.Doc(chatSnapshot.Ref.ID))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		response.New(http.StatusInternalServerError).Err("failed to clear the chat in a transaction").Send(c)
		return
	}

	// Respond with success
	response.New(http.StatusOK).Send(c)
}
