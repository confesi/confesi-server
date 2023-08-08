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

	mux.POST("/token-anon", h.handleSetTokenAnon)

	// any firebase users
	mux.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.AllFbUsers, []string{}, middleware.NeedsAll)
	})

	mux.POST("/token-uid", h.handleSetTokenWithUid)
	mux.DELETE("/token", h.handleRemoveToken)
	mux.GET("/topic-prefs", h.handleGetTopicPrefs)
	mux.PUT("/topic-prefs", h.handleSetTopicPrefs)
}
