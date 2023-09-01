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
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	if post.UserID == "" {
		// Handle the situation where there is no matching non-hidden post for the given ID.
		response.New(http.StatusBadRequest).Err("post not found or is hidden").Send(c)
		return
	}

	if post.UserID == token.UID {
		response.New(http.StatusBadRequest).Err("you can't DM yourself").Send(c)
		return
	}

	uniqueID := generateUniqueID(token.UID, post.UserID, post.ID.ToString())
	docRef := h.fb.FirestoreClient.Collection("rooms").Doc(uniqueID)

	room := db.Room{
		UserCreator: token.UID,
		UserOther:   post.UserID,
		PostID:      post.ID.ToInt(),
		Name:        uuid.New().String(), // temp name before (if) a participant changes it
		LastMsg:     time.Now().UTC(),
	}

	// Try to create the room with the unique ID
	_, err = docRef.Create(c, room)
	if err != nil {
		// If the document already exists, Firestore will return an error
		if status.Code(err) == codes.AlreadyExists {
			response.New(http.StatusBadRequest).Err("room with this combination already exists").Send(c)
			return
		}

		// For other errors:
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
		WithMsg(builders.AdminSendNotificationNoti("title", "body")).
		WithData(map[string]string{}).
		Send()

	// Send a success response
	response.New(http.StatusOK).Send(c)
}

// Generate a unique ID from the given values
func generateUniqueID(userCreator, userOther, postID string) string {
	combined := userCreator + ":" + userOther + ":" + postID
	return combined
}
