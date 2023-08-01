package logger

import (
	"fmt"
	"os"
	"time"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

func StdErr(m error) {
	now := time.Now()
	year, month, date := now.UTC().Date()
	hour := now.UTC().Hour()
	minute := now.UTC().Minute()

	// Time, Handler, Message/Error, Ip, UID, UserContext
	//month-date-year hour:minute | handler | ip | uid | error/message

	str := fmt.Sprintf("%v-%v-%v %v:%v | %s \n", year, int(month), date, hour, minute, m.Error())
	os.Stderr.Write([]byte(str))
}

func ResErr(m error, ctx *gin.Context, statusCode int) {
	now := time.Now()
	year, month, date := now.UTC().Date()
	hour := now.UTC().Hour()
	minute := now.UTC().Minute()

	endpoint := ctx.Request.URL.Path

	ip := ctx.ClientIP()

	uid := ""
	user, _ := ctx.Get("user")
	token, _ := user.(*auth.Token)
	if token != nil {
		uid = token.UID
	}

	// Time, Handler, Message/Error, Ip, UID, UserContext
	//month-date-year hour:minute | statusCode | handler | ip | uid | error/message

	str := fmt.Sprintf("%v-%v-%v %v:%v | %v | %s | %s | %s | %s \n", year, int(month), date, hour, minute, statusCode, endpoint, ip, uid, m.Error())
	os.Stderr.Write([]byte(str))
}
