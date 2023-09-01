package dms

import (
	"confesi/config/builders"
	"confesi/db"
	"confesi/lib/encryption"
	fcm "confesi/lib/firebase_cloud_messaging"
	"confesi/lib/response"
	"confesi/lib/utils"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *handler) handleCreateRoom(c *gin.Context) {

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c) // Fix for the undefined serverError
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

	// find user_id from the post by id if it's not hidden
	var post db.Post
	err = h.db.
		Where("id = ?", unmaskedPostId).
		Where("hidden = ?", false).
		// Where("chat_post = ?", true). // todo: add chat post
		First(&post).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// Handle the error appropriately
		response.New(http.StatusInternalServerError).Err("failed to fetch post data, or invalid post").Send(c)
		return
	} else if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		// Handle the situation where there is no matching non-hidden post for the given ID.
		response.New(http.StatusBadRequest).Err("post not found").Send(c)
		return
	}

	if post.UserID == token.UID {
		response.New(http.StatusBadRequest).Err("you can't DM yourself").Send(c)
		return
	}

	// Check if room with these users and postID already exists
	// Get an iterator of matching documents
	iter := h.fb.FirestoreClient.Collection("rooms").
		Where("u_1", "==", token.UID).
		Where("u_2", "==", post.UserID).
		Where("post_id", "==", post.ID.ToInt()).
		Documents(c)

	// Check if any document exists
	doc, err := iter.Next()
	if err == nil && doc != nil {
		response.New(http.StatusBadRequest).Err("room with this combination already exists").Send(c)
		return
	}

	room := db.Room{
		U1:      token.UID,
		U2:      post.UserID,
		PostID:  post.ID.ToInt(),
		Name:    "New chat",
		LastMsg: time.Now().UTC(),
	}

	// Create the room with Firestore's automatic ID generation
	_, _, err = h.fb.FirestoreClient.Collection("rooms").Add(c, room)
	if err != nil {
		fmt.Println(err)
		response.New(http.StatusInternalServerError).Err("failed to create room").Send(c)
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

	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	go fcm.New(h.fb.MsgClient).
		ToTokens(tokens).
		WithMsg(builders.NewRoomCreatedNoti()).
		WithData(map[string]string{}).
		Send()

	// Send a success response
	response.New(http.StatusOK).Send(c)
}
