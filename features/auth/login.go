package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleLogin(c *gin.Context) {
	c.Status(http.StatusOK)
}
