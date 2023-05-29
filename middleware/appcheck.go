package middleware

import (
	"confesi/lib"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AppCheck(c *gin.Context) {
	requestToken := c.GetHeader("X-AppCheck-Token")
	if requestToken != token {
		c.AbortWithStatus(http.StatusForbidden)
		url := c.Request.URL.String()
		lib.StdErr(errors.New("unauthorized request to: " + url))
		return
	}

	c.Next()
}
