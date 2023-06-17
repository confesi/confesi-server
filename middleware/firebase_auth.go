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
//
// Allows for additional checks to see if a user possess some specified roles.
func UsersOnly(c *gin.Context, auth *auth.Client, allowedUser AllowedUser, roles []string) {
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
		if profileCreated, ok := token.Claims["profile_created"].(bool); !ok {
			// registered user without postgres profile since the claim isn't created till after their account gets saved to postgres
			response.New(http.StatusUnauthorized).Err("registered user without profile").Send(c)
			return
		} else if profileCreated {
			// registered user with postgres profile, now we check if they have the required roles
			rolesClaim, ok := token.Claims["roles"]
			if !ok {
				response.New(http.StatusUnauthorized).Err("roles field doesn't exist in claims").Send(c)
				return
			}

			rolesInterfaceSlice, ok := rolesClaim.([]interface{})
			if !ok {
				response.New(http.StatusUnauthorized).Err("invalid roles field in claims").Send(c)
				return
			}

			parsedRoles := make([]string, len(rolesInterfaceSlice))
			for i, role := range rolesInterfaceSlice {
				if strRole, ok := role.(string); ok {
					parsedRoles[i] = strRole
				} else {
					response.New(http.StatusUnauthorized).Err("invalid role value in roles field").Send(c)
					return
				}
			}

			// check if all the required roles exist in the parsed roles
			for _, requiredRole := range roles {
				found := false
				for _, role := range parsedRoles {
					if requiredRole == role {
						found = true
						break
					}
				}
				if !found {
					response.New(http.StatusUnauthorized).Err("invalid role").Send(c)
					return
				}
			}

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
