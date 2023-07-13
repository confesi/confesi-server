package notifications

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

func (h *handler) handleSubToPriv(c *gin.Context) {
	// validate request
	var req validation.FcmPrivQuery
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// todo: FK to ensure there's a match between either valid sub type or name of watchd uni ? Or just check if valid topic?
	// todo: DOES the current SQL migration #23 enforce the FK constraint? or just say there is an FK?

	fcmPriv := db.FcmPriv{
		UserID: token.UID,
	}
	if req.ContentType == "post" {
		fcmPriv.PostID = req.ContentID
	} else if req.ContentType == "comment" {
		fcmPriv.CommentID = req.ContentID
	} else {
		// should never happen with validated struct, but to be defensive
		response.New(http.StatusBadRequest).Err(fmt.Sprintf("invalid content type")).Send(c)
		return
	}

	err = h.db.
		Where("user_id = ? AND comment_id = ? AND post_id = ?", token.UID, fcmPriv.CommentID, fcmPriv.PostID).
		FirstOrCreate(&fcmPriv).
		Error
	if err != nil {
		var pgErr *pgconn.PgError
		// Gorm doesn't properly handle duplicate errors: https://github.com/go-gorm/gorm/issues/4037
		if ok := errors.As(err, &pgErr); !ok {
			// if it's not a PostgreSQL error, return a generic server error
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
		switch pgErr.Code {
		case "23503": // foreign key constraint violation
			response.New(http.StatusBadRequest).Err("invalid content id").Send(c) // aka, you provided an invalid post/comment id
			return
		default:
			// some other postgreSQL error
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	response.New(http.StatusOK).Send(c)
}
