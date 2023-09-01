package dms

import (
	"confesi/db"
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

	// validate the json body from request
	var req validation.AddChat
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	// get user token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// if token.UID is not a subet of room_id, return error
	//
	// this works because the room_id is a combination of the two user ids with the post ID messaged over
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

	// Define the transaction function
	err = h.fb.FirestoreClient.RunTransaction(c, func(ctx context.Context, tx *firestore.Transaction) error {
		// Add the chat to Firestore
		chatsCollectionRef := h.fb.FirestoreClient.Collection("chats")
		_, _, err := chatsCollectionRef.Add(c, chat)
		if err != nil {
			return err
		}

		// Update last_msg field in the room document
		roomRef := h.fb.FirestoreClient.Collection("rooms").Doc(req.RoomID)
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

	// Send a success response
	response.New(http.StatusOK).Send(c)
}
