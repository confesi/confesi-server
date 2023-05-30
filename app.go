package main

import (
	"confesi/features/auth"
	"confesi/lib/fire"
	"confesi/middleware"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

var port string
var app *fire.FirebaseApp

func init() {
	port = os.Getenv("PORT")
	if port == "" {
		fmt.Println("PORT env not found, using default 8080")
		port = "8080"
	}

	// Init Firebase app
	err := fire.InitFirebase("firebase-secrets.json")
	if err != nil {
		// if we can't init firebase, we have an unrecoverable error
		log.Fatal("Error initializing Firebase app: ", err)
	}
}

func main() {
	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(middleware.AppCheck)

	// Version 1 api group, alongside core middleware
	api := r.Group("/api/v1")
	api.Use(middleware.Cors)
	api.Use(gin.Recovery())

	// Separate handler groups
	auth.Router(api.Group("/auth"))

	r.Run(fmt.Sprintf(":%s", port))
}
