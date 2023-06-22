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

	// anybody
	mux.GET("/hottest", h.handleGetHottest)
	mux.GET("/post", h.handleGetPostById)

	// any firebase user
	anyFirebaseUserRoutes := mux.Group("")
	anyFirebaseUserRoutes.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.AllFbUsers, []string{})
	})
	anyFirebaseUserRoutes.GET("/posts", h.handleGetPosts)
	anyFirebaseUserRoutes.DELETE("/purge", h.handlePurgePostsCache)

	// only registered firebase users
	registeredFirebaseUserRoutes := mux.Group("")
	registeredFirebaseUserRoutes.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers, []string{})
	})
	registeredFirebaseUserRoutes.POST("/create", h.handleCreate)
}
