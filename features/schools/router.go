package schools

import (
	"confesi/db"
	"confesi/lib/fire"
	"confesi/middleware"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	serverError = errors.New("server error")
	invalidId   = errors.New("invalid id")
)

type handler struct {
	*gorm.DB
	fb *fire.FirebaseApp
}

func Router(r *gin.RouterGroup) {
	h := handler{db.New(), fire.New()}

	r.GET("/", h.getSchools)

	// protected route
	protectedRoutes := r.Group("")
	protectedRoutes.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers, []string{})
	})
	protectedRoutes.POST("/watch", h.handleWatchSchool)
	protectedRoutes.DELETE("/unwatch", h.handleUnwatchSchool)
	protectedRoutes.GET("/watched", h.handleGetWatchedSchools)
}
