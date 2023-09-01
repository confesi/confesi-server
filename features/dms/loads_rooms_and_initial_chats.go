package dms

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"net/http"
	"sort"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
)

const InitialChatCount = 10 // number of initial chats per room to load

type LoadedRoom struct {
	db.Room
	ID      string    `json:"id"`
	Chats   []db.Chat `json:"chats"`
	UserNum int       `json:"user_num"`
}

func (h *handler) handleLoadRoomsAndInitialChats(c *gin.Context) {
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		return
	}

	processedRooms := make(map[string]bool) // map to track processed rooms by their document ID
	loadedRooms := make([]LoadedRoom, 0)

	// Helper function to process rooms
	processRooms := func(iter *firestore.DocumentIterator, userNum int) {
		for {
			doc, err := iter.Next()
			if err != nil {
				break
			}

			// If the room is already processed, skip
			if _, exists := processedRooms[doc.Ref.ID]; exists {
				continue
			}
			processedRooms[doc.Ref.ID] = true

			var room db.Room
			if err := doc.DataTo(&room); err != nil {
				response.New(http.StatusInternalServerError).Err("error decoding room data").Send(c)
				return
			}

			// Fetch initial chats
			chatsIter := h.fb.FirestoreClient.Collection("chats").
				Where("room_id", "==", doc.Ref.ID).
				OrderBy("date", firestore.Desc).
				Limit(InitialChatCount).
				Documents(c)

			chats := make([]db.Chat, 0)
			for {
				chatDoc, chatErr := chatsIter.Next()
				if chatErr != nil {
					break
				}
				var chat db.Chat
				if chatErr := chatDoc.DataTo(&chat); chatErr != nil {
					response.New(http.StatusInternalServerError).Err("error decoding chat data").Send(c)
					return
				}
				chats = append(chats, chat)
			}

			// Reverse chats for chronological order
			sort.SliceStable(chats, func(i, j int) bool {
				return chats[i].Date.Before(chats[j].Date)
			})

			loadedRooms = append(loadedRooms, LoadedRoom{
				Room:    room,
				ID:      doc.Ref.ID,
				Chats:   chats,
				UserNum: userNum,
			})
		}
	}

	// Get rooms by u_1
	iterU1 := h.fb.FirestoreClient.Collection("rooms").Where("u_1", "==", token.UID).Documents(c)
	processRooms(iterU1, 1)

	// Get rooms by u_2
	iterU2 := h.fb.FirestoreClient.Collection("rooms").Where("u_2", "==", token.UID).Documents(c)
	processRooms(iterU2, 2)

	response.New(http.StatusOK).Val(loadedRooms).Send(c)
}
