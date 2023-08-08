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
	notFound    = errors.New("not found")
)

type PostDetail struct {
	db.Post  `json:"post"`
	UserVote int      `json:"user_vote"`
	Owner    bool     `json:"owner"`
	Emojis   []string `json:"emojis" gorm:"-"`
}

type handler struct {
	db    *gorm.DB
	fb    *fire.FirebaseApp
	redis *redis.Client
}

func Router(mux *gin.RouterGroup) {
	h := handler{db: db.New(), fb: fire.New(), redis: cache.New()}

	// anybody
	mux.GET("/sentiment", h.sentimentAnaylsis)

	// any firebase user
	anyFirebaseUserRoutes := mux.Group("")
	anyFirebaseUserRoutes.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.AllFbUsers, []string{}, middleware.NeedsAll)
	})
	anyFirebaseUserRoutes.GET("/posts", h.handleGetPosts)
	anyFirebaseUserRoutes.DELETE("/purge", h.handlePurgePostsCache)
	anyFirebaseUserRoutes.GET("/hottest", h.handleGetHottest)
	anyFirebaseUserRoutes.GET("/post", h.handleGetPostById)

	// only registered firebase users
	registeredFirebaseUserRoutes := mux.Group("")
	registeredFirebaseUserRoutes.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers, []string{}, middleware.NeedsAll)
	})
	registeredFirebaseUserRoutes.PATCH("/hide", h.handleHidePost)
	registeredFirebaseUserRoutes.POST("/create", h.handleCreate)
	registeredFirebaseUserRoutes.GET("/your-posts", h.handleGetYourPosts)
	registeredFirebaseUserRoutes.PATCH("/edit", h.handleEditPost)

}
