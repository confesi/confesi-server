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
	invalidInput     = errors.New("invalid content")
	notFound         = errors.New("not found")
)

type CommentDetail struct {
	db.Comment `json:"comment"`
	UserVote   int  `json:"user_vote" gorm:"column:user_vote"`
	Owner      bool `json:"owner"`
	Saved      bool `json:"saved"`
	Reported   bool `json:"reported"`
}

type handler struct {
	db    *gorm.DB
	fb    *fire.FirebaseApp
	redis *redis.Client
}

func Router(mux *gin.RouterGroup) {
	h := handler{db: db.New(), fb: fire.New(), redis: cache.New()}

	// any firebase user
	anyFirebaseUserRoutes := mux.Group("")
	anyFirebaseUserRoutes.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.AllFbUsers, []string{})
	})
	anyFirebaseUserRoutes.GET("/roots", h.handleGetComments)
	anyFirebaseUserRoutes.GET("/replies", h.handleGetReplies)
	anyFirebaseUserRoutes.GET("/comment", h.handleGetCommentById)
	anyFirebaseUserRoutes.DELETE("/purge", h.handlePurgeCommentsCache)

	// registered firebase users only
	registeredFirebaseUserOnly := mux.Group("")
	registeredFirebaseUserOnly.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers, []string{})
	})
	registeredFirebaseUserOnly.POST("/create", h.handleCreate)
	registeredFirebaseUserOnly.PATCH("/hide", h.handleHideComment)
	registeredFirebaseUserOnly.GET("/your-comments", h.handleGetYourComments)
	registeredFirebaseUserOnly.PATCH("/edit", h.handleEditComment)
}
