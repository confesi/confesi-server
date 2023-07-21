package auth

import (
	"confesi/db"
	"confesi/lib/fire"
	"confesi/middleware"
	"errors"

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

	// registered firebase users only
	registeredFirebaseUserRoutes := mux.Group("")
	registeredFirebaseUserRoutes.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers, []string{})
	})

	registeredFirebaseUserRoutes.PATCH("/update-email", h.handleUpdateEmail)
	registeredFirebaseUserRoutes.POST("/resend-email-verification", h.handleResendEmailVerification)
}
