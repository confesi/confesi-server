package auth

import (
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

func (h *handler) handleLogin(c *gin.Context, authClient *auth.Client) {
	c.Status(http.StatusOK)
}
