package comments

import (
	"confesi/db"
	"confesi/lib/fire"
	"confesi/middleware"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// todo: for deleted comments, they should STILL GET RETURNED so they don't screw up the tree-structure, but just have their `content` set to "[removed]"

var (
	serverError      = errors.New("server error")
	threadDepthError = errors.New("thread depth error")
)

type handler struct {
	db *gorm.DB
	fb *fire.FirebaseApp
}

func Router(mux *gin.RouterGroup) {
	h := handler{db: db.New(), fb: fire.New()}

	// only allow registered users to create a post
	mux.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers, []string{})
	})
	mux.POST("/create", h.handleCreate)
	mux.PATCH("/hide", h.handleHideComment)
}
