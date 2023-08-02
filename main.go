package main

import (
	"confesi/features/admin"
	"confesi/features/auth"
	"confesi/features/comments"
	"confesi/features/feedback"
	hideLog "confesi/features/hide_log"
	"confesi/features/notifications"
	"confesi/features/posts"
	"confesi/features/reports"
	"confesi/features/saves"
	"confesi/features/schools"
	"confesi/features/user"
	"confesi/features/votes"
	"confesi/lib/cronJobs/clearExpiredFcmTokens"
	"confesi/lib/cronJobs/dailyHottestPosts"
	"confesi/middleware"
	"fmt"
	"os"

	swaggerFiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware

	_ "confesi/docs" // docs file

	"github.com/gin-gonic/gin"
)

var port string
var publicDocAccess string

func init() {
	port = os.Getenv("PORT")
	if port == "" {
		fmt.Println("PORT env not found, using default 8080")
		port = "8080"
	}
	publicDocAccess = os.Getenv("PUBLIC_DOC_ACCESS")
	if publicDocAccess != "true" && publicDocAccess != "false" {
		panic("`PUBLIC_DOC_ACCESS env not found or invalid (true or false)")
	}
}

// @title           Confesi dev-only API docs
// @version         1.0

// @host      localhost:8080
// @BasePath  /api/v1

// @externalDocs.description  GitHub
// @externalDocs.url          https://github.com/mattrltrent/confesi-server
func main() {
	r := gin.Default()

	if publicDocAccess == "true" {
		// Swagger docs
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Gin settings
	r.SetTrustedProxies(nil)

	// Version 1 api group
	api := r.Group("/api/v1")
	//! Not used for now, since: https://github.com/firebase/firebase-admin-go/issues/572
	// api.Use(func(c *gin.Context) {
	// 	middleware.FirebaseAppCheck(c, fire.New().AppCheck)
	// })
	api.Use(middleware.RateLimit)
	api.Use(middleware.LatLong)
	api.Use(middleware.Cors)
	api.Use(gin.Recovery())

	// Separate handler groups
	comments.Router(api.Group("/comments"))
	auth.Router(api.Group("/auth"))
	posts.Router(api.Group("/posts"))
	votes.Router(api.Group("/votes"))
	schools.Router(api.Group("/schools"))
	saves.Router(api.Group("/saves"))
	admin.Router(api.Group("/admin"))
	feedback.Router(api.Group("/feedback"))
	notifications.Router(api.Group("/notifications"))
	user.Router(api.Group("/user"))
	reports.Router(api.Group("/reports"))
	hideLog.Router(api.Group("/hide-log"))

	// Start the CRON job scheduler
	dailyHottestPosts.StartDailyHottestPostsCronJob()
	clearExpiredFcmTokens.StartClearExpiredFcmTokensCronJob()

	// Start the server
	r.Run(fmt.Sprintf(":%s", port))
}
