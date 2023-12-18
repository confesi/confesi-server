package saves

import (
	"confesi/db"
	"confesi/lib/encryption"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

func (h *handler) unsaveContent(c *gin.Context, token *auth.Token, req validation.SaveContentDetails, unmaskedId uint) error {
	var err error
	if req.ContentType == "post" {
		savedPost := db.SavedPost{
			UserID: token.UID,
			PostID: db.EncryptedID{Val: unmaskedId},
		}
		err = h.db.Delete(&savedPost, "user_id = ? AND post_id = ?", savedPost.UserID, savedPost.PostID).Error
	} else if req.ContentType == "comment" {
		savedComment := db.SavedComment{
			UserID:    token.UID,
			CommentID: db.EncryptedID{Val: unmaskedId},
		}
		err = h.db.Delete(&savedComment, "user_id = ? AND comment_id = ?", savedComment.UserID, savedComment.CommentID).Error
	} else {
		return serverError
	}
	if err != nil {
		return serverError
	}
	return nil
}

func (h *handler) handleUnsave(c *gin.Context) {
	// extract request
	var req validation.SaveContentDetails
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	unmaskedId, err := encryption.Unmask(req.ContentID)
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid id").Send(c)
		return
	}

	err = h.unsaveContent(c, token, req, unmaskedId)
	if err != nil {
		// all returned errors are just general client-facing "server errors"
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		return
	}

	// if all goes well, send 200
	response.New(http.StatusOK).Send(c)
}
