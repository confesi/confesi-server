package middleware

import (
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

type AllowedUser string

const (
	AllFbUsers        AllowedUser = "all_fb_users"
	RegisteredFbUsers AllowedUser = "registered_fb_users"
)

// Only allows valid Firebase users to pass through.
//
// Sets the authenticated Firebase user id to the context as `user`
// iff the user is anon or registered.
func UsersOnly(c *gin.Context, auth *auth.Client, allowedUser AllowedUser) {
	idToken := c.GetHeader("Authorization")

	// ensure the token is valid, aka, there is some valid user
	token, err := auth.VerifyIDToken(c, idToken)
	if err != nil {
		// no fb user at all, or malformed token
		c.Status(http.StatusUnauthorized)
		return
	}

	if allowedUser == AllFbUsers {
		c.Set("user", token)
		c.Next()
		return
	}

	// assume user isn't registered to start
	isRegistered := false

	if firebaseClaims, exists := token.Claims["firebase"]; exists { // TODO: do I have to key into "firebase"?
		if firebaseMap, ok := firebaseClaims.(map[string]interface{}); ok {
			if provider, providerExists := firebaseMap["sign_in_provider"]; providerExists { // TODO: is "sign_in_provider" the right string?
				if providerStr, ok := provider.(string); ok {
					isRegistered = providerStr == "email" // TODO: is "email" the right string?
				}
			}
		}
	}

	if isRegistered {
		c.Set("user", token)
		c.Next()
	} else {
		// registered users only! you're an anon user.
		c.Status(http.StatusUnauthorized)
	}

}
