package user

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
)

type handler struct {
	db    *gorm.DB
	fb    *fire.FirebaseApp
	redis *redis.Client
}

func Router(mux *gin.RouterGroup) {
	h := handler{db: db.New(), fb: fire.New(), redis: cache.New()}

	// registered firebase users only
	registeredFirebaseUsersOnly := mux.Group("")
	registeredFirebaseUsersOnly.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers, []string{})
	})
	registeredFirebaseUsersOnly.GET("/user", h.handleGetUser)
	registeredFirebaseUsersOnly.GET("/user-standing", h.handleGetUserStanding)

	registeredFirebaseUsersOnly.DELETE("/faculty", h.handleClearFaculty)
	registeredFirebaseUsersOnly.PATCH("/faculty", h.handleSetFaculty)

	registeredFirebaseUsersOnly.DELETE("/year-of-study", h.handleClearYearOfStudy)
	registeredFirebaseUsersOnly.PATCH("/year-of-study", h.handleSetYearOfStudy)

	registeredFirebaseUsersOnly.PATCH("/school", h.handleSetSchool)

	registeredFirebaseUsersOnly.GET("/user-stats", h.handleGetUserStats)
	registeredFirebaseUsersOnly.GET("/awards", h.handleGetAwards)
}
