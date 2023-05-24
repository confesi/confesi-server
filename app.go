package main

import (
	"confesi/features/auth"
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

	auth.Router(r.Group("/auth"))

	r.Run(fmt.Sprintf(":%s", port))
}
