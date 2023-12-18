package main

import (
	"confesi/handlers/admin"
	"confesi/handlers/auth"
	"confesi/handlers/comments"
	"confesi/handlers/dms"
	"confesi/handlers/drafts"
	"confesi/handlers/feedback"
	hideLog "confesi/handlers/hide_log"
	"confesi/handlers/notifications"
	"confesi/handlers/posts"
	"confesi/handlers/public"
	"confesi/handlers/reports"
	"confesi/handlers/saves"
	"confesi/handlers/schools"
	"confesi/handlers/user"
	"confesi/handlers/votes"
	"confesi/lib/cronJobs/clearExpiredFcmTokens"
	"confesi/lib/cronJobs/cronNotifications"
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

	// api middlewares
	api.Use(middleware.RateLimit)
	api.Use(middleware.Cors)
	api.Use(gin.Recovery())
	api.Use(middleware.OptionalProfanityCensor)

	// Static assets
	static := r.Group("/")
	static.Use(middleware.RateLimit)
	static.Use(middleware.Cors)
	static.Use(gin.Recovery())

	// Static middlewares

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
	drafts.Router(api.Group("/drafts"))
	dms.Router(api.Group("/dms"))

	// static assets like images, css, html, etc. so from root path
	public.Router(static.Group("/"))

	// Start the CRON job scheduler
	dailyHottestPosts.StartDailyHottestPostsCronJob()
	clearExpiredFcmTokens.StartClearExpiredFcmTokensCronJob()
	cronNotifications.StartDailyHottestPostsCronJob()

	// Start the server
	r.Run(fmt.Sprintf(":%s", port))
}
