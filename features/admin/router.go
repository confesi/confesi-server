package admin

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
	serverError  = errors.New("server error")
	notFound     = errors.New("not found")
	invalidValue = errors.New("invalid value")
)

type handler struct {
	db    *gorm.DB
	fb    *fire.FirebaseApp
	redis *redis.Client
}

func Router(mux *gin.RouterGroup) {
	h := handler{db: db.New(), fb: fire.New(), redis: cache.New()}
	mux.Use(func(c *gin.Context) {
		//! ADMINS ONLY FOR THESE ROUTES. VERY IMPORTANT.
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers, []string{"admin"})
	})
	mux.PATCH("/user-standing", h.handleUserStanding)
	mux.POST("/daily-hottest-cron", h.handleManuallyTriggerDailyHottestCron)
	mux.POST("/expire-fcm-tokens", h.handleManuallyTriggerClearExpiredFcmTokens)
	mux.GET("/crons", h.handleGetDailyHottestCrons)
	mux.GET("/feedback", h.handleListFeedback)
	mux.GET("/feedback/:feedbackID", h.handleFeedbackID)
	mux.GET("/report", h.handleGetReportById)
	mux.GET("/reports", h.handleGetReports)
	mux.PATCH("/hide", h.handleHideContent)
	mux.PATCH("/reviewed-by-mod", h.handleReviewContentByMod)
	mux.GET("/comments-by-report", h.handleGetRankedCommentsByReport)
	mux.GET("/posts-by-report", h.handleGetRankedPostsByReport)
}
