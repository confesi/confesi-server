package comments

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
	serverError      = errors.New("server error")
	threadDepthError = errors.New("thread depth error")
)

type CommentDetail struct {
	db.Comment `json:"comment"`
	UserVote   int `json:"user_vote"`
}

type handler struct {
	db    *gorm.DB
	fb    *fire.FirebaseApp
	redis *redis.Client
}

func Router(mux *gin.RouterGroup) {
	h := handler{db: db.New(), fb: fire.New(), redis: cache.New()}

	// any user
	mux.GET("/comment", h.handleGetCommentById)

	// any firebase user
	anyFirebaseUserRoutes := mux.Group("")
	anyFirebaseUserRoutes.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.AllFbUsers, []string{})
	})
	anyFirebaseUserRoutes.GET("/roots", h.handleGetComments)
	anyFirebaseUserRoutes.GET("/replies", h.handleGetReplies)
	anyFirebaseUserRoutes.DELETE("/purge", h.handlePurgeCommentsCache)

	// only allow registered users to create a post
	mux.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers, []string{})
	})
	mux.POST("/create", h.handleCreate)
	mux.PATCH("/hide", h.handleHideComment)
}