package dms

import (
	"confesi/config/builders"
	"confesi/db"
	"confesi/lib/encryption"
	fcm "confesi/lib/firebase_cloud_messaging"
	"confesi/lib/response"
	"confesi/lib/utils"
	"context"
	"errors"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"gorm.io/gorm"
)

func (h *handler) handleCreateRoom(c *gin.Context) {
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		return
	}

	post_id := c.Query("post-id")
	if post_id == "" {
		response.New(http.StatusBadRequest).Err("post-id query param required").Send(c)
		return
	}

	// unmask
	unmaskedPostId, err := encryption.Unmask(post_id)
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid masked post id").Send(c)
		return
	}

	var post db.Post
	err = h.db.
		Where("id = ?", unmaskedPostId).
		Where("hidden = ?", false).
		First(&post).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.New(http.StatusBadRequest).Err("post not found").Send(c)
		} else {
			response.New(http.StatusInternalServerError).Err("failed to fetch post data").Send(c)
		}
		return
	}

	if post.UserID == token.UID {
		response.New(http.StatusBadRequest).Err("you can't DM yourself").Send(c)
		return
	}

	// Check if a room involving the token user for this postID already exists
	_, err = h.fb.FirestoreClient.Collection("rooms").
		Where("post_id", "==", post.ID.ToInt()).
		Where("user_id", "==", token.UID).
		Documents(c).Next()
	if err != nil && err != iterator.Done {
		response.New(http.StatusInternalServerError).Err("failed to check for existing rooms").Send(c)
		return
	}
	tokenUserRoomExists := err != iterator.Done

	// Check if a room involving the post user for this postID already exists
	_, err = h.fb.FirestoreClient.Collection("rooms").
		Where("post_id", "==", post.ID.ToInt()).
		Where("user_id", "==", post.UserID).
		Documents(c).Next()
	if err != nil && err != iterator.Done {
		response.New(http.StatusInternalServerError).Err("failed to check for existing rooms").Send(c)
		return
	}
	postUserRoomExists := err != iterator.Done

	if post.UserID == token.UID {
		response.New(http.StatusBadRequest).Err("you can't DM yourself").Send(c)
		return
	}

	if tokenUserRoomExists || postUserRoomExists {
		response.New(http.StatusBadRequest).Err("room with this combination already exists").Send(c)
		return
	}

	// Generate a unique room_id using UUID
	roomID := uuid.New().String()

	// Creating two rooms: one for the token user and another for the post user
	currentUserRoom := db.Room{
		UserID:     token.UID,
		PostID:     post.ID.ToInt(),
		Name:       "New chat",
		LastMsg:    time.Now().UTC(),
		UserNumber: 1,
		RoomID:     roomID,
	}

	postUserRoom := db.Room{
		UserID:     post.UserID,
		PostID:     post.ID.ToInt(),
		Name:       "New chat",
		LastMsg:    time.Now().UTC(),
		UserNumber: 2,
		RoomID:     roomID,
	}

	// Use Firestore transactions for atomic operations
	err = h.fb.FirestoreClient.RunTransaction(c, func(ctx context.Context, tx *firestore.Transaction) error {
		// Add rooms to Firestore
		if err := tx.Set(h.fb.FirestoreClient.Collection("rooms").NewDoc(), currentUserRoom); err != nil {
			return err
		}
		if err := tx.Set(h.fb.FirestoreClient.Collection("rooms").NewDoc(), postUserRoom); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		response.New(http.StatusInternalServerError).Err("failed to create rooms").Send(c)
		return
	}

	// Obtain FCM tokens for the affected other user
	var tokens []string
	err = h.db.
		Table("fcm_tokens").
		Select("fcm_tokens.token").
		Joins("JOIN users ON users.id = fcm_tokens.user_id").
		Where("users.id = ?", post.UserID).
		Pluck("fcm_tokens.token", &tokens).
		Error

	// ignore errors

	go fcm.New(h.fb.MsgClient).
		ToTokens(tokens).
		WithMsg(builders.NewRoomCreatedNoti()).
		WithData(builders.NewRoomCreatedData(roomID)).
		Send()

	response.New(http.StatusOK).Send(c)
}
