package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleRegister(c *gin.Context) {
	c.Status(http.StatusOK)
}
