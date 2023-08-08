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

type reportDetail struct {
	db.Report   `gorm:"embedded"`
	ContentType string `json:"content_type" gorm:"-"`
}

type fetchResults struct {
	Reports []reportDetail `json:"reports"`
	Next    *int64         `json:"next"`
}

func Router(mux *gin.RouterGroup) {
	h := handler{db: db.New(), fb: fire.New(), redis: cache.New()}
	mux.Use(func(c *gin.Context) {
		//! ADMINS & MODS ONLY FOR THESE ROUTES. VERY IMPORTANT. ANY EDITS TO THIS SHOULD RAISE RED FLAGS.
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers, []string{"admin", "mod", "s_mod"}, middleware.NeedsOne)
		//! ADMINS & MODS ONLY FOR THESE ROUTES. VERY IMPORTANT. ANY EDITS TO THIS SHOULD RAISE RED FLAGS.
	})

	// ? User Standing
	mux.PATCH("/user-standing", h.handleSetUserStanding)
	mux.GET("/user-standing", h.handleGetUserStanding) // Any Moderator can get user standing

	// ? Content

	mux.PATCH("/reviewed-by-mod", h.handleReviewContentByMod)
	mux.PATCH("/hide", h.handleHideContent) // Only posts are uni specific, comments are not

	// ? Reports
	mux.GET("/report", h.handleGetReportById)                          // TODO: Reports do not have school_id
	mux.GET("/reports", h.handleGetReports)                            // TODO: Reports do not have school_id
	mux.GET("/comments-by-report", h.handleGetRankedCommentsByReport)  // TODO: Reports do not have school_id
	mux.GET("/posts-by-report", h.handleGetRankedPostsByReport)        // TODO: Reports do not have school_id
	mux.GET("/reports-for-comment", h.handleFetchReportForCommentById) // TODO: Reports do not have school_id
	mux.GET("/reports-for-post", h.handleFetchReportForPostById)       // TODO: Reports do not have school_id

	mux.Use(func(c *gin.Context) {
		//! ADMINS ONLY FOR THESE ROUTES. VERY IMPORTANT. ANY EDITS TO THIS SHOULD RAISE RED FLAGS.
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers, []string{"admin"}, middleware.NeedsAll)
		//! ADMINS ONLY FOR THESE ROUTES. VERY IMPORTANT. ANY EDITS TO THIS SHOULD RAISE RED FLAGS.
	})

	// ? Cron Jobs
	mux.POST("/daily-hottest-cron", h.handleManuallyTriggerDailyHottestCron)
	mux.POST("/expire-fcm-tokens", h.handleManuallyTriggerClearExpiredFcmTokens)
	mux.GET("/crons", h.handleGetDailyHottestCrons)

	// ? Feedback
	mux.GET("/feedback", h.handleListFeedback)
	mux.GET("/feedback/:feedbackID", h.handleFeedbackID)

	// ? User
	mux.PATCH("/set-user-role", h.handleSetUserRole)

}

func getUserRoles(c *gin.Context) (*middleware.UserRoleTypes, error) {
	userRoleTypesContext, ok := c.Get("userRoleTypes")
	if !ok {
		return nil, errors.New("userRoleTypes not found in context")
	}
	roleTypes, err := userRoleTypesContext.(middleware.UserRoleTypes)
	if !err {
		return nil, serverError
	}
	return &roleTypes, nil
}
