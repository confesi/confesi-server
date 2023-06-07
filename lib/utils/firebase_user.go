package utils

import (
	"errors"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

func UserFromContext(c *gin.Context) (*auth.Token, error) {
	user, ok := c.Get("user")
	if !ok {
		return nil, errors.New("invalid user")
	}

	token, ok := user.(*auth.Token)
	if !ok {
		return nil, errors.New("invalid user")
	}
	return token, nil
}
