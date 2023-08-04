package notifications

import (
	"confesi/db"
	fcm "confesi/lib/firebase_cloud_messaging"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

func (h *handler) handleSetTokenAnon(c *gin.Context) {

	// validate request
	var req validation.FcmTokenQuery
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	// is validate FCM token?
	if !fcm.IsValidFcmToken(h.fb.MsgClient, req.Token) {
		response.New(http.StatusBadRequest).Err(fcm.InvalidFcmTokenError.Error()).Send(c)
		return
	}

	// start a transaction
	tx := h.db.Begin()

	// if something goes ary, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}()

	fcmToken := db.FcmToken{
		Token: req.Token,
	}

	// Update the existing record if it exists
	result := tx.Model(&fcmToken).
		Where("token = ?", req.Token).
		Updates(map[string]interface{}{
			"updated_at": time.Now(), // set new value for updated_at
		})
	if err != nil {
		var pgErr *pgconn.PgError
		// Gorm doesn't properly handle duplicate errors: https://github.com/go-gorm/gorm/issues/4037
		if ok := errors.As(err, &pgErr); !ok {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
		switch pgErr.Code {
		case "23505": // duplicate key value violates unique constraint
			err = tx.Commit().Error
			if err != nil {
				tx.Rollback()
				response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
				return
			}
			response.New(http.StatusCreated).Send(c)
			return
		default:
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}
	if result.RowsAffected == 0 {
		// Record not found, create a new one
		fcmToken.UpdatedAt.Time = time.Now()
		err = tx.Create(&fcmToken).Error
		if err != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	// if all goes well, respond with a 201 & commit the transaction
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	response.New(http.StatusCreated).Send(c)
}
