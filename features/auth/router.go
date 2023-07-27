package auth

import (
	"confesi/config"
	"confesi/db"
	"confesi/lib/fire"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/middleware"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	serverError       = errors.New("server error")
	errorSendingEmail = errors.New("error sending email")
)

type handler struct {
	db *gorm.DB
	fb *fire.FirebaseApp
}

func Router(mux *gin.RouterGroup) {
	h := handler{db: db.New(), fb: fire.New()}

	// any user
	mux.POST("/register", h.handleRegister)

	anyFirebaseUserWithRateLimiting := mux.Group("")
	anyFirebaseUserWithRateLimiting.Use(func(c *gin.Context) {
		middleware.RoutedRateLimit(c, 3, time.Hour, config.RedisEmailRateLimitingRouteKey, c.ClientIP(), "too many emails sent")
	})
	anyFirebaseUserWithRateLimiting.POST("/send-password-reset-email", h.handleSendPasswordResetEmail)

	anyFirebaseUser := mux.Group("")
	anyFirebaseUser.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.AllFbUsers, []string{})
	})
	// protect against email spam by UID
	anyFirebaseUser.Use(func(c *gin.Context) {
		token, err := utils.UserTokenFromContext(c)
		if err != nil {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
		middleware.RoutedRateLimit(c, 3, time.Hour, config.RedisEmailRateLimitingRouteKey, token.UID, "too many emails sent")
	})
	anyFirebaseUser.POST("/resend-verification-email", h.handleResendEmailVerification)

	// registered firebase users only
	registeredFirebaseUserRoutes := mux.Group("")
	registeredFirebaseUserRoutes.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers, []string{})
	})
}
