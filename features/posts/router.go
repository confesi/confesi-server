package posts

import (
	"confesi/db"
	"confesi/lib/cache"
	"confesi/lib/fire"
	"errors"

	"confesi/middleware"

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

	// allow any user to get the hottest posts
	mux.GET("/hottest", h.handleGetHottest)
	mux.GET("/post", h.handleGetPostById)

	anyFirebaseUserRoutes := mux.Group("")
	anyFirebaseUserRoutes.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.AllFbUsers)
	})
	anyFirebaseUserRoutes.GET("/posts", h.handleGetPosts)
	anyFirebaseUserRoutes.DELETE("/purge", h.handlePurgePostsCache)

	// only allow registered users to create a post
	registeredFirebaseUserRoutes := mux.Group("")
	registeredFirebaseUserRoutes.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers)
	})
	registeredFirebaseUserRoutes.POST("/create", h.handleCreate)
}
