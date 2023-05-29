package auth

import (
	"confesi/lib/response"
	"log"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

// Example creating a Firebase user
func (h *handler) handleRegister(c *gin.Context) {
	params := (&auth.UserToCreate{}).
		Email("user99@example.com").
		Password("examplePassword").
		Disabled(false)

	user, err := h.fb.AuthClient.CreateUser(c, params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error creating user": err})
		return
	}
	log.Printf("Successfully created user: %v\n", user)
	response.New(http.StatusOK).
		Val(gin.H{"email": user.Email}).
		Send(c)
}
