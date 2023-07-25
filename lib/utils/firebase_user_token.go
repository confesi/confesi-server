package utils

import (
	"confesi/lib/crypto"
	"encoding/base64"
	"errors"
	"os"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

var (
	userAD []byte
)

func init() {
	userAD = []byte(os.Getenv("USER_KEY"))
	if len(userAD) == 0 {
		panic("`USER_KEY` not set")
	}
}

func UserTokenFromContext(c *gin.Context) (*auth.Token, error) {
	user, ok := c.Get("user")
	if !ok {
		return nil, errors.New("invalid user")
	}

	token, ok := user.(*auth.Token)
	if !ok {
		return nil, errors.New("invalid user")
	}
	ciphertext, err := crypto.Cipher([]byte(token.UID), userAD)
	if err != nil {
		return nil, err
	}
	token.UID = base64.StdEncoding.EncodeToString(ciphertext)
	return token, nil
}
