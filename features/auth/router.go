package auth

import (
	"confesi/db"
	"confesi/lib/fire"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type handler struct {
	db *gorm.DB
	fb *fire.FirebaseApp
}

func Router(mux *gin.RouterGroup) {
	h := handler{db: db.New(), fb: fire.New()}

	mux.POST("/register", h.handleRegister)
}
