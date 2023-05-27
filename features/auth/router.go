package auth

import (
	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

func Router(mux *gin.RouterGroup, authClient *auth.Client) {
	mux.POST("/login", func(c *gin.Context) {
		handleLogin(c, authClient)
	})
	mux.POST("/register", func(c *gin.Context) {
		handleRegister(c, authClient)
	})
}
