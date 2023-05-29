package middleware

import (
	"confesi/lib"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AppCheck(c *gin.Context) {
	requestToken := c.GetHeader("X-AppCheck-Token")
	if requestToken != token {
		c.AbortWithStatus(http.StatusForbidden)
		url := c.Request.URL.String()
		ip := c.ClientIP()
		str := fmt.Errorf("unauthorized request:\nfrom %s\nto: %s", ip, url)
		lib.StdErr(str)
		return
	}

	c.Next()
}
