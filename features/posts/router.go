package posts

import (
	"confesi/db"
	"confesi/lib/cache"
	"confesi/lib/fire"

	"confesi/middleware"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type handler struct {
	db    *gorm.DB
	fb    *fire.FirebaseApp
	redis *redis.Client
}

func Router(mux *gin.RouterGroup) {
	h := handler{db: db.New(), fb: fire.New(), redis: cache.New()}

	// allow any user to get the hottest posts
	mux.GET("/hottest", h.handleGetHottest)
	mux.GET("/post", h.handleGetPostById)

	// only allow registered users to create a post
	protectedRoutes := mux.Group("")
	protectedRoutes.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers)
	})
	protectedRoutes.POST("/create", h.handleCreate)
}
