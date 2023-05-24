package auth

import (
	"github.com/gin-gonic/gin"
)

func Router(mux *gin.RouterGroup) {
	mux.POST("/login", handleLogin)
	mux.POST("/register", handleRegister)
}
