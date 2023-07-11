package notifications

import (
	"confesi/db"
	"confesi/lib/fire"
	"confesi/middleware"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	serverError     = errors.New("server error")
	failedSubbing   = errors.New("failed subbing")
	failedUnsubbing = errors.New("failed unsubbing")
)

type handler struct {
	db *gorm.DB
	fb *fire.FirebaseApp
}

func Router(mux *gin.RouterGroup) {
	h := handler{db: db.New(), fb: fire.New()}

	// all firebase users
	mux.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.AllFbUsers, []string{})
	})

	mux.PATCH("/sub", h.handleSubToTopic)
	mux.PATCH("/unsub", h.handleUnsubToTopic)
	mux.POST("/token", h.handleSetToken)
	mux.DELETE("/token", h.handleRemoveToken)
}
