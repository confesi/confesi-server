package dms

import (
	"confesi/config/builders"
	fcm "confesi/lib/firebase_cloud_messaging"
	"confesi/lib/response"
	"confesi/lib/utils"
	"context"
	"fmt"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *handler) handleDeleteChat(c *gin.Context) {
	// Authenticate the user and obtain their token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		fmt.Println("1")
		fmt.Println(err)
		response.New(http.StatusInternalServerError).Err("failed to get user token").Send(c)
		return
	}

	// Extract chatID from the request
	chatID := c.Query("id")
	if chatID == "" {
		response.New(http.StatusBadRequest).Err("id parameter required").Send(c)
		return
	}

	fmt.Println(chatID)

	// Fetch the chat details first for later notification use
	chatRef := h.fb.FirestoreClient.Collection("chats").Doc(chatID)
	chatSnapshot, err := chatRef.Get(c)
	if err != nil {
		fmt.Println("3")
		fmt.Println(err)
		response.New(http.StatusInternalServerError).Err("failed to fetch chat message").Send(c)
		return
	}

	roomID, ok := chatSnapshot.Data()["room_id"].(string)
	if !ok {
		fmt.Println("4")
		fmt.Println(err)
		response.New(http.StatusInternalServerError).Err("failed to extract room ID from chat message").Send(c)
		return
	}

	err = h.fb.FirestoreClient.RunTransaction(c, func(ctx context.Context, tx *firestore.Transaction) error {
		// Delete the specific chat message within transaction
		err := tx.Delete(chatRef)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to delete chat message: %v", err)
		}
		return nil
	})

	if err != nil {
		fmt.Println("2")
		fmt.Println(err)
		response.New(http.StatusInternalServerError).Err("transaction error").Send(c)
		return
	}

	// Get all users of the room
	otherUserRoomQuery := h.fb.FirestoreClient.Collection("rooms").
		Where("room_id", "==", roomID)

	docsIterator := otherUserRoomQuery.Documents(c)

	var otherUserIDs []string

	for {
		otherUserRoomSnapshot, err := docsIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Println("5")
			fmt.Println(err)
			response.New(http.StatusInternalServerError).Err("error checking chat room membership").Send(c)
			return
		}

		// Check if this userID is different from the authenticated user's UID
		if uid, ok := otherUserRoomSnapshot.Data()["user_id"].(string); ok && uid != token.UID {
			otherUserIDs = append(otherUserIDs, uid)
		}
	}

	// If there are no other users in the room
	if len(otherUserIDs) == 0 {
		response.New(http.StatusBadRequest).Err("No other users found in the specified chat room").Send(c)
		return
	}

	// Send notifications to all other users
	for _, otherUserUid := range otherUserIDs {
		// Spin up light-weight thread to send FCM message
		var tokens []string
		err = h.db.
			Table("fcm_tokens").
			Select("fcm_tokens.token").
			Joins("JOIN users ON users.id = fcm_tokens.user_id").
			Where("users.id = ?", otherUserUid).
			Pluck("fcm_tokens.token", &tokens).
			Error
		if err != nil {
			fmt.Println("7")
			fmt.Println(err)
			response.New(http.StatusInternalServerError).Err("failed to get FCM tokens").Send(c)
			return
		}

		// (don't handle error case since it's not necessary)
		go fcm.New(h.fb.MsgClient).
			ToTokens(tokens).
			WithMsg(builders.DeletedChatNoti()).
			WithData(builders.DeletedChatData()).
			Send()
	}

	// Respond with success
	response.New(http.StatusOK).Send(c)
}
