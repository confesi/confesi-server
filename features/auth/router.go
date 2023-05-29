package auth

import (
	"confesi/db"
	"confesi/lib/fire"
	"confesi/middleware"
	"time"

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
		middleware.RateLimit(c, 10, time.Second*15)
	})

	mux.POST("/login", func(c *gin.Context) {
		h.handleLogin(c)
	})
	mux.POST("/register", func(c *gin.Context) {
		h.handleRegister(c)
	})
}
