package middleware

import (
	"confesi/lib/response"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

type AllowedUser string

const (
	AllFbUsers        AllowedUser = "all_fb_users"        // anon or registered users (aka, we don't care about their account)
	RegisteredFbUsers AllowedUser = "registered_fb_users" // only registered users who have a postgres profile (fully created account)
)

// Only allows valid Firebase users to pass through.
//
// Sets the authenticated Firebase user id to the context as `user`
// iff the user is anon or registered.
func UsersOnly(c *gin.Context, auth *auth.Client, allowedUser AllowedUser) {
	idToken := c.GetHeader("Authorization")
	if len(idToken) < 8 || idToken[:7] != "Bearer " {
		response.New(http.StatusUnauthorized).Err("malformed Authorization header").Send(c)
		return
	}
	tokenValue := idToken[7:] // extract the token value by removing the "Bearer " prefix

	// ensure the token is valid, aka, there is some valid user
	token, err := auth.VerifyIDToken(c, tokenValue)
	if err != nil {
		// no Firebase user or malformed token
		response.New(http.StatusUnauthorized).Err("invalid token").Send(c)
		return
	}

	// anon or registered user; the resource is okay with taking either
	// and it doesn't care if the potentially registered user has a postgres profile
	if allowedUser == AllFbUsers {
		c.Set("user", token)
		c.Next()
		return
	}

	if token.Firebase.SignInProvider == "password" {
		var profileCreated bool
		var ok bool
		if profileCreated, ok = token.Claims["profile_created"].(bool); !ok {
			// registered user without postgres profile since the claim isn't created till after their account gets saved to postgres
			response.New(http.StatusUnauthorized).Err("registered user without profile").Send(c)
			return
		} else if profileCreated {
			// registered user with postgres profile
			c.Set("user", token)
			c.Next()
			return
		} else {
			// registered user without postgres profile (handling the future case where the claim at "profile_created" is turned back to false for some reason)
			response.New(http.StatusUnauthorized).Err("registered user without profile").Send(c)
			return
		}

	} else {
		// anon user (but resource requires registered user)
		response.New(http.StatusUnauthorized).Err("registered users only").Send(c)
	}

}

// Example token. See how it looks on jwt.io.
// eyJhbGciOiJSUzI1NiIsImtpZCI6IjJkM2E0YTllYjY0OTk0YzUxM2YyYzhlMGMwMTY1MzEzN2U5NTg3Y2EiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vY29uZmVzaS1zZXJ2ZXItZGV2IiwiYXVkIjoiY29uZmVzaS1zZXJ2ZXItZGV2IiwiYXV0aF90aW1lIjoxNjg1MjUwNTI4LCJ1c2VyX2lkIjoiVXhHUTZkc2ZnN1k3NkVJcTNUaG55YU45cVlFMyIsInN1YiI6IlV4R1E2ZHNmZzdZNzZFSXEzVGhueWFOOXFZRTMiLCJpYXQiOjE2ODUyNTA1MjgsImV4cCI6MTY4NTI1NDEyOCwiZW1haWwiOiJjbGllbnQzQGV4YW1wbGUuY29tIiwiZW1haWxfdmVyaWZpZWQiOmZhbHNlLCJmaXJlYmFzZSI6eyJpZGVudGl0aWVzIjp7ImVtYWlsIjpbImNsaWVudDNAZXhhbXBsZS5jb20iXX0sInNpZ25faW5fcHJvdmlkZXIiOiJwYXNzd29yZCJ9fQ.pDahYSHsT8W_H_x6sle_Yb7HZt4mwqggT_2JWuOjny2M05isbcNghIBLnKnvbzU8hqkGvqz5sZs021AQL9pzA0JDWhkNjvzCKwNi06cYPyosfcoDiG3izg6P4NxJSbLYzKdgEU1jyKaKX3EfsQ5EZo5Ag_ErHfELLKMPhHlwvbV4Cf-KdlWSBKsi1Bt9vzr5LdXbhvwmsg35jpajUI-PvsWu8yS8k0-gqn9hub4yZhslPRZgs8Xr0VRjrMwVyQ13fFNVfGUmIT3CBZ1foMJ7Y3csBhrDl-qF4SrHSoo6uMFg-7lpz_jX_x7XntL3cB4NEno6trTSy7NIduNLwgdpnw
