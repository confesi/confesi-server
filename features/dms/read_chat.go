package dms

import (
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
)

func (h *handler) handleReadChat(c *gin.Context) {
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("failed to get user token").Send(c)
		return
	}

	// Extract request data
	var req validation.ReadRoomRequest
	err = utils.New(c).Validate(&req)
	if err != nil {
		response.New(http.StatusBadRequest).Err("failed to validate request").Send(c)
		return
	}

	// update the "read" from the room in firestore to current time
	roomQuery := h.fb.FirestoreClient.Collection("rooms").
		Where("room_id", "==", req.RoomID).
		Where("user_id", "==", token.UID)

	// Getting the documents matching the query
	docs, err := roomQuery.Documents(c).GetAll()
	if err != nil {
		response.New(http.StatusInternalServerError).Err("failed to fetch room data").Send(c)
		return
	}

	// If no documents were found, send a 404 response
	if len(docs) == 0 {
		response.New(http.StatusNotFound).Err("no room found").Send(c)
		return
	}

	// Update "read" in each document found (though we expect only one match)
	for _, doc := range docs {
		_, err = doc.Ref.Set(c, map[string]interface{}{
			"read": time.Now().UTC(),
		}, firestore.MergeAll)
		if err != nil {
			response.New(http.StatusInternalServerError).Err("failed to update last read message time").Send(c)
			return
		}
	}

	// Send a success response
	response.New(http.StatusOK).Send(c)
}
