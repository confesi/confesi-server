package dms

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

func (h *handler) handleUpdateChatName(c *gin.Context) {
	// Get user token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("failed to get user token").Send(c)
		return
	}

	// Extract request data
	var req validation.UpdateChatName
	err = utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	roomQuery := h.fb.FirestoreClient.Collection("rooms").
		Where("room_id", "==", req.RoomID).
		Where("user_id", "==", token.UID)

	roomSnapshotIterator := roomQuery.Documents(c)
	roomSnapshot, err := roomSnapshotIterator.Next()

	if err == iterator.Done {
		response.New(http.StatusBadRequest).Err("room not found with given criteria").Send(c)
		return
	} else if err != nil {
		response.New(http.StatusInternalServerError).Err("error querying room").Send(c)
		return
	}

	// Process roomSnapshot as needed
	var room db.Room
	if err := roomSnapshot.DataTo(&room); err != nil {
		response.New(http.StatusInternalServerError).Err("failed decoding room data").Send(c)
		return
	}

	if token.UID != room.UserID {
		response.New(http.StatusBadRequest).Err("user is not part of the room").Send(c)
		return
	}

	// Update the room name in Firestore
	_, err = roomSnapshot.Ref.Update(c, []firestore.Update{
		{Path: "name", Value: req.NewName},
	})

	if err != nil {
		response.New(http.StatusInternalServerError).Err("failed to update room name").Send(c)
		return
	}

	response.New(http.StatusOK).Send(c)
}
