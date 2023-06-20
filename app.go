package main

import (
	"confesi/features/admin"
	"confesi/features/auth"
	"confesi/features/feedback"
	"confesi/features/posts"
	"confesi/features/schools"
	"confesi/features/votes"
	"confesi/middleware"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

var port string

func init() {
	port = os.Getenv("PORT")
	if port == "" {
		fmt.Println("PORT env not found, using default 8080")
		port = "8080"
	}
}

func main() {
	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(middleware.AppCheck)

	// Version 1 api group, alongside core middleware
	api := r.Group("/api/v1")
	api.Use(middleware.RateLimit)
	api.Use(middleware.Cors)
	api.Use(gin.Recovery())

	// Separate handler groups
	auth.Router(api.Group("/auth"))
	posts.Router(api.Group("/posts"))
	votes.Router(api.Group("/votes"))
	schools.Router(api.Group("/schools"))
	admin.Router(api.Group("/admin"))
	feedback.Router(api.Group("/feedback"))

	r.Run(fmt.Sprintf(":%s", port))
}
