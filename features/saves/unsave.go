package saves

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"fmt"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func (h *handler) unsaveContent(c *gin.Context, token *auth.Token, req validation.SaveContentDetails) error {
	var err error
	if req.ContentType == "post" {
		savedPost := db.SavedPost{
			UserID: token.UID,
			PostID: req.ContentID,
		}
		err = h.db.Delete(&savedPost, "user_id = ? AND post_id = ?", savedPost.UserID, savedPost.PostID).Error
	} else if req.ContentType == "comment" {
		savedComment := db.SavedComment{
			UserID:    token.UID,
			CommentID: req.ContentID,
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

	// create validator
	validator := validator.New()

	// create a binding instance with the validator, check if json valid, if so, deserialize into req
	binding := &validation.DefaultBinding{
		Validator: validator,
	}
	if err := binding.Bind(c.Request, &req); err != nil {
		response.New(http.StatusBadRequest).Err(fmt.Sprintf("failed validation: %v", err)).Send(c)
		return
	}
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	err = h.unsaveContent(c, token, req)
	if err != nil {
		// all returned errors are just general client-facing "server errors"
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		return
	}

	// if all goes well, send 200
	response.New(http.StatusOK).Send(c)
}
