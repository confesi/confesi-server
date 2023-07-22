package middleware

import (
	"confesi/lib/logger"
	"confesi/lib/response"
	"fmt"
	"net/http"

	"firebase.google.com/go/v4/appcheck"
	"github.com/gin-gonic/gin"
)

func FirebaseAppCheck(c *gin.Context, appCheck *appcheck.Client) {
	appCheckToken := c.GetHeader("X-Firebase-AppCheck")

	_, err := appCheck.VerifyToken(appCheckToken)
	if err != nil {
		url := c.Request.URL.String()
		ip := c.ClientIP()
		logger.StdErr(fmt.Errorf("unauthorized firebase appcheck request:\nfrom %s\nto: %s", ip, url))
		response.New(http.StatusUnauthorized).Err("fails appcheck").Send(c)
		return
	}

	c.Next()
}
