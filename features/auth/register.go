package auth

import (
	"context"
	"log"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

// Example creating a Firebase user
func handleRegister(c *gin.Context, authClient *auth.Client) {
	params := (&auth.UserToCreate{}).
		Email("user@example.com").
		Password("examplePassword").
		Disabled(false)

	user, err := authClient.CreateUser(context.Background(), params)
	if err != nil {
		log.Fatalf("error creating user: %v\n", err)
	}
	log.Printf("Successfully created user: %v\n", user)
	c.JSON(http.StatusOK, gin.H{"email": user.Email})
}
