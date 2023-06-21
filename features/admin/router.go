package admin

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
	mux.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers) //TODO: transition to the middleware auth such that only admins can hit this
	})
	mux.PATCH("/user-standing", h.handleUserStanding)
	mux.POST("/daily-hottest-cron", h.handleManuallyTriggerDailyHottestCron)
}
