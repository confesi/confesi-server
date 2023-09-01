package dms

import (
	"confesi/config"
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type LoadedRoom struct {
	db.Room
	ID      string    `json:"id"`
	Chats   []db.Chat `json:"chats"`
	UserNum int       `json:"user_num"`
}

const (
	roomsCacheExpiry = 24 * time.Hour
	InitialChatCount = 10
)

func (h *handler) handleLoadRoomsAndInitialChats(c *gin.Context) {
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		return
	}

	var req validation.FetchRooms
	if err := utils.New(c).Validate(&req); err != nil {
		return
	}

	cacheKey := createCacheKey(token.UID, req.SessionKey)

	if req.PurgeCache && h.redis.Del(c, cacheKey).Err() != nil {
		response.New(http.StatusInternalServerError).Err("error purging cache").Send(c)
		return
	}

	cachedRooms, err := getCachedRooms(h, c, cacheKey)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("error accessing cache").Send(c)
		return
	}

	loadedRooms, err := fetchRoomsFromFirestore(h, c, token.UID, cachedRooms)
	if err != nil {
		if err.Error() == "no more items in iterator" {
			response.New(http.StatusInternalServerError).Err("error fetching rooms from firestore: no more items in iterator").Send(c)
			return
		}
		response.New(http.StatusInternalServerError).Err("error fetching rooms from firestore: " + err.Error()).Send(c)
		return
	}

	if h.redis.Expire(c, cacheKey, roomsCacheExpiry).Err() != nil {
		response.New(http.StatusInternalServerError).Err("error setting cache expiry").Send(c)
		return
	}

	response.New(http.StatusOK).Val(loadedRooms).Send(c)
}

func createCacheKey(uid, sessionKey string) string {
	return config.RedisRoomsCache + ":" + uid + ":" + sessionKey
}

func getCachedRooms(h *handler, c *gin.Context, cacheKey string) (map[string]bool, error) {
	cachedIDs, err := h.redis.SMembers(c, cacheKey).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	roomMap := make(map[string]bool)
	for _, id := range cachedIDs {
		roomMap[id] = true
	}
	return roomMap, nil
}

func fetchRoomsFromFirestore(h *handler, c *gin.Context, uid string, cachedRooms map[string]bool) ([]LoadedRoom, error) {
	var loadedRooms []LoadedRoom

	processRooms := func(iter *firestore.DocumentIterator, userNum int) error {
		for {
			doc, err := iter.Next()

			if err != nil {
				if err.Error() == "no more items in iterator" {
					break
				}
				return err
			}

			roomID := doc.Ref.ID
			if _, alreadyProcessed := cachedRooms[roomID]; alreadyProcessed {
				continue
			}

			var room db.Room
			if err := doc.DataTo(&room); err != nil {
				return err
			}

			chats, err := fetchInitialChats(h, c, roomID)
			if err != nil {
				return err
			}

			loadedRooms = append(loadedRooms, LoadedRoom{
				Room:    room,
				ID:      roomID,
				Chats:   chats,
				UserNum: userNum,
			})

			h.redis.SAdd(c, roomID)
		}
		return nil
	}

	if err := processRooms(h.fb.FirestoreClient.Collection("rooms").Where("u_1", "==", uid).Documents(c), 1); err != nil {
		return nil, err
	}
	if err := processRooms(h.fb.FirestoreClient.Collection("rooms").Where("u_2", "==", uid).Documents(c), 2); err != nil {
		return nil, err
	}
	return loadedRooms, nil
}

func fetchInitialChats(h *handler, c *gin.Context, roomID string) ([]db.Chat, error) {
	chatsIter := h.fb.FirestoreClient.Collection("chats").
		Where("room_id", "==", roomID).
		OrderBy("date", firestore.Desc).
		Limit(InitialChatCount).
		Documents(c)

	var chats []db.Chat
	for {
		chatDoc, err := chatsIter.Next()

		if err != nil {
			if err.Error() == "no more items in iterator" {
				break
			}
			return nil, err
		}

		var chat db.Chat
		if err := chatDoc.DataTo(&chat); err != nil {
			return nil, err
		}
		chats = append(chats, chat)
	}

	return chats, nil
}
