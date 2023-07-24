package middleware

import (
	"confesi/db"
	"confesi/lib/fire"
	"confesi/lib/response"
	"confesi/lib/validation"
	"errors"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
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

	// if they are an email-password user (like all our registered users)
	if token.Firebase.SignInProvider == "password" {
		// if their email is NOT verified, then send back not verified
		// todo: UNCOMMENT IN REAL IMPLEMENTATION; COMMENTED OUT FOR TESTING
		if !token.Claims["email_verified"].(bool) {
			response.New(http.StatusUnauthorized).Val("email not verified").Send(c)
			return
		}
		// todo: add check for `disabled` users to block them, too.
		// todo: UNCOMMENT IN REAL IMPLEMENTATION; COMMENTED OUT FOR TESTING
		if profileCreated, ok := token.Claims["sync"].(bool); !ok {
			// registered user without postgres profile (handling the future case where the claim at "sync" is turned back to false for some reason)
			err := RetrySyncPostgresAccountCreation(c, token)
			if err != nil {
				response.New(http.StatusUnauthorized).Err("non-synced account").Send(c)
				return
			}
		} else if profileCreated {
			// registered user with postgres profile, now we check if they have the required roles
			var rolesClaim interface{}
			if rolesClaim, ok = token.Claims["roles"]; !ok {
				response.New(http.StatusUnauthorized).Err("roles field doesn't exist in claims").Send(c)
				return
			}
			var rolesInterfaceSlice []interface{}
			if rolesInterfaceSlice, ok = rolesClaim.([]interface{}); !ok {
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
			// registered user without postgres profile (handling the future case where the claim at "sync" is turned back to false for some reason)
			err := RetrySyncPostgresAccountCreation(c, token)
			if err != nil {
				response.New(http.StatusUnauthorized).Err("non-synced account").Send(c)
				return
			}
		}
	} else {
		// anon user (but resource requires registered user)
		response.New(http.StatusUnauthorized).Err("registered users only").Send(c)
	}
}

var (
	errorExtractingEmailDomain = errors.New("error extracting domain from email")
	domainDoesntBelongToSchool = errors.New("domain doesn't belong to school")
	serverError                = errors.New("server error")
)

func RetrySyncPostgresAccountCreation(c *gin.Context, token *auth.Token) error {
	// get the user's email from their token
	userEmail := token.Claims["email"].(string)

	// create postgres user
	user := db.User{}
	user.ID = token.UID

	// extract domain from user's email
	domain, err := validation.ExtractEmailDomain(userEmail)
	if err != nil {
		return errorExtractingEmailDomain
	}

	// get connections
	dbConn := db.New()
	authClient := fire.New().AuthClient

	// start a transaction
	tx := dbConn.Begin()
	// if something goes ary, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			return
		}
	}()

	// check if user's email is valid
	var school db.School
	err = tx.Select("id").Where("domain = ?", domain).First(&school).Error
	if err != nil {
		tx.Rollback()
		return domainDoesntBelongToSchool
	}

	// else, add the email to the user
	user.SchoolID = school.ID

	err = tx.Create(&user).Error
	if err != nil {
		var pgErr *pgconn.PgError
		// Gorm doesn't properly handle duplicate errors: https://github.com/go-gorm/gorm/issues/4037
		if ok := errors.As(err, &pgErr); !ok {
			// if it's not a PostgreSQL error, return a generic server error
			tx.Rollback()
			return serverError
		}
		switch pgErr.Code {
		case "23505": // duplicate key value violates unique constraint
			// dont do anything; user exists!
		default:
			// some other postgreSQL error
			tx.Rollback()
			return serverError
		}
	}

	// update custom claims on token
	err = authClient.SetCustomUserClaims(c, token.UID, map[string]interface{}{
		"sync":  true,
		"roles": []string{}, //! default users have no roles, VERY IMPORTANT
	})
	// don't catch the error! if it fails, we'll just catch it next time

	// commit the transaction
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return serverError
	}

	return nil
}
