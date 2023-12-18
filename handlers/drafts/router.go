package drafts

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

type DraftDetail struct {
	db.Draft `json:"draft"`
}

type handler struct {
	db    *gorm.DB
	fb    *fire.FirebaseApp
	redis *redis.Client
}

func Router(mux *gin.RouterGroup) {
	h := handler{db: db.New(), fb: fire.New(), redis: cache.New()}

	// only registered firebase users
	registeredFirebaseUserRoutes := mux.Group("")
	registeredFirebaseUserRoutes.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers, []string{})
	})
	registeredFirebaseUserRoutes.POST("/create", h.handleCreate)
	registeredFirebaseUserRoutes.GET("/your-drafts", h.handleGetYourDrafts)
	registeredFirebaseUserRoutes.PATCH("/edit", h.handleEditDraft)
	registeredFirebaseUserRoutes.DELETE("/delete", h.handleDeleteDraft)
}
