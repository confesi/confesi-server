package auth

import (
	"log"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

// Example creating a Firebase user
func handleRegister(c *gin.Context, authClient *auth.Client) {
	params := (&auth.UserToCreate{}).
		Email("user99@example.com").
		Password("examplePassword").
		Disabled(false)

	user, err := authClient.CreateUser(c, params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error creating user": err})
		return
	}
	log.Printf("Successfully created user: %v\n", user)
	c.JSON(http.StatusOK, gin.H{"email": user.Email})
}
