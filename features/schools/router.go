package schools

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
	serverError = errors.New("server error")
	invalidId   = errors.New("invalid id")
)

type handler struct {
	*gorm.DB
	fb    *fire.FirebaseApp
	redis *redis.Client
}

func Router(r *gin.RouterGroup) {
	h := handler{db.New(), fire.New(), cache.New()}

	// any user
	r.GET("/", h.getSchools)
	r.GET("/random", h.handleGetRandomSchool)

	// any firebase user
	anyFirebaseUser := r.Group("")
	anyFirebaseUser.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.AllFbUsers, []string{})
	})
	anyFirebaseUser.GET("/rank", h.handleGetRankedSchools)
	anyFirebaseUser.DELETE("/purge", h.handlePurgeRankedSchoolsCache)

	// only registered firebase users
	registeredFirebaseUsersOnly := r.Group("")
	registeredFirebaseUsersOnly.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers, []string{})
	})
	registeredFirebaseUsersOnly.POST("/watch", h.handleWatchSchool)
	registeredFirebaseUsersOnly.DELETE("/unwatch", h.handleUnwatchSchool)
	registeredFirebaseUsersOnly.GET("/watched", h.handleGetWatchedSchools)
}
