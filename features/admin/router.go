package admin

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
)

type handler struct {
	db *gorm.DB
	fb *fire.FirebaseApp
}

func Router(mux *gin.RouterGroup) {
	h := handler{db: db.New(), fb: fire.New()}
	mux.Use(func(c *gin.Context) {
		//! ADMINS ONLY FOR THESE ROUTES. VERY IMPORTANT.
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers, []string{"admin"})
	})
	mux.PATCH("/user-standing", h.handleUserStanding)
	mux.POST("/daily-hottest-cron", h.handleManuallyTriggerDailyHottestCron)
	mux.GET("/daily-hottest-crons", h.handleGetDailyHottestCrons)
	mux.GET("/feedback", h.handleListFeedback)
	mux.GET("/feedback/:feedbackID", h.handleFeedbackID)
}
