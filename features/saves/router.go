package saves

import (
	"confesi/db"
	"confesi/lib/fire"
	"confesi/middleware"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type handler struct {
	db *gorm.DB
	fb *fire.FirebaseApp
}

var (
	serverError = errors.New("server error")
	invalidId   = errors.New("invalid id")
)

func Router(mux *gin.RouterGroup) {
	h := handler{db: db.New(), fb: fire.New()}

	// only allow registered users to save content
	mux.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers, []string{})
	})

	mux.POST("/save", h.handleSave)
	mux.DELETE("/unsave", h.handleUnsave)
	// these two endpoints have some code duplication, but it
	// appears to be cleaner than to try to combine them and have to
	// deal with union types
	mux.GET("/posts", h.handleGetPosts)
	mux.GET("/comments", h.handleGetComments)
}
