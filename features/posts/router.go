package posts

import (
	"confesi/db"
	"confesi/lib/fire"
	"confesi/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type handler struct {
	db *gorm.DB
	fb *fire.FirebaseApp
}

func Router(mux *gin.RouterGroup) {
	h := handler{db: db.New(), fb: fire.New()}

	// allow any user to get the hottest posts
	mux.GET("/hottest", h.handleGetHottest)

	// only allow registered users to create a post
	protectedRoutes := mux.Group("")
	protectedRoutes.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers)
	})
	protectedRoutes.POST("/create", h.handleCreate)

}
