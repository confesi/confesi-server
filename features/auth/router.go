package auth

import (
	"confesi/db"
	"confesi/middleware"
	"time"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type handler struct {
	db       *gorm.DB
	firebase any // TODO: firebase embedded struct
}

func Router(mux *gin.RouterGroup, authClient *auth.Client) {
	h := handler{db: db.New(), firebase: nil}

	mux.Use(func(c *gin.Context) {
		middleware.RateLimit(c, 10, time.Minute)
	})

	mux.POST("/login", func(c *gin.Context) {
		h.handleLogin(c, authClient)
	})
	mux.POST("/register", func(c *gin.Context) {
		handleRegister(c, authClient)
	})
}
