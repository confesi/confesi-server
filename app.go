package main

import (
	"confesi/config"
	"confesi/features/auth"
	"confesi/middleware"
	"fmt"
	"log"
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

	// Init firebase app
	app, err := config.InitFirebase("firebase-secrets.json")
	if err != nil {
		log.Fatal("Error initializing Firebase app: ", err)
	}

	// Version 1 api group, alongside core middleware
	api := r.Group("/api/v1")
	api.Use(middleware.Cors)
	api.Use(gin.Recovery())

	// Separate handler groups
	auth.Router(api.Group("/auth"), app.AuthClient)

	r.Run(fmt.Sprintf(":%s", port))
}
