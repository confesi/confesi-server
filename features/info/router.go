package info

import (
	"confesi/db"
	"confesi/lib/cache"
	"confesi/lib/fire"
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

	mux.GET("/", h.handleIndex)
}
