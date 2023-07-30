package utils

import (
	"errors"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

var (
	ErrInvalidUser = errors.New("invalid user")
)

func UserTokenFromContext(c *gin.Context) (*auth.Token, error) {
	user, ok := c.Get("user")
	if !ok {
		return nil, ErrInvalidUser
	}

	token, ok := user.(*auth.Token)
	if !ok {
		return nil, ErrInvalidUser
	}
	return token, nil
}
