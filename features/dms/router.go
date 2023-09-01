package dms

import (
	"confesi/db"
	"confesi/lib/cache"
	"confesi/lib/fire"
	"confesi/middleware"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"gorm.io/gorm"
)

var (
	serverError = errors.New("server error")
)

type handler struct {
	db    *gorm.DB
	fb    *fire.FirebaseApp
	redis *redis.Client
}

func Router(mux *gin.RouterGroup) {
	h := handler{db: db.New(), fb: fire.New(), redis: cache.New()}

	// only registered firebase users
	mux.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers, []string{})
	})

	mux.POST("/rooms", h.handleCreateRoom)
	mux.POST("/chat", h.handleAddChat)
	mux.GET("/rooms", h.handleLoadRoomsAndInitialChats)
}
