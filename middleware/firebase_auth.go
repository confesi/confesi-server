package middleware

import (
	"confesi/db"
	"confesi/lib/fire"
	"confesi/lib/response"
	"confesi/lib/validation"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRoleTypes struct {
	Admin      bool  `json:"admin"`       // can do anything
	GlobalMod  bool  `json:"global_mod"`  // can do mod actions for any school
	SchoolMods []int `json:"school_mods"` // can do mod actions for specific schools
}

type AllowedUser string

const (
	AllFbUsers        AllowedUser = "all_fb_users"        // anon or registered users (aka, we don't care about their account)
	RegisteredFbUsers AllowedUser = "registered_fb_users" // only registered users who have a postgres profile (fully created account)
)

type RoleRequirements string

const (
	NeedsOne RoleRequirements = "strict"  // every role listed must be present in the user's roles
	NeedsAll RoleRequirements = "relaxed" // at least one of the roles listed must be present in the user's roles
)

// Only allows valid Firebase users to pass through.
//
// Sets the authenticated Firebase user id to the context as `user`
// iff the user is anon or registered.
//
// Allows for additional checks to see if a user possess some specified roles.
func UsersOnly(c *gin.Context, auth *auth.Client, allowedUser AllowedUser, roles []string, roleRequirements RoleRequirements) {
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
		if !token.Claims["email_verified"].(bool) {
			response.New(http.StatusUnauthorized).Val("email not verified").Send(c)
			return
		}
		// disabled users will get locked out after at most 1 hour
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
			// create user role types
			userRoleTypes, err := createUserRoleTypes(rolesInterfaceSlice)
			if err != nil {
				response.New(http.StatusInternalServerError).Err("server error").Send(c)
				return
			}

			// check if all the required roles exist in the parsed roles
			{
				found := false
				for _, requiredRole := range roles {
					if roleRequirements == NeedsAll {
						found = false
					}
					for _, role := range parsedRoles {
						if requiredRole == role {
							found = true
							break
						}
					}
					if !found && roleRequirements == NeedsAll {
						response.New(http.StatusUnauthorized).Err("invalid role").Send(c)
						return
					}
				}
				if !found && roleRequirements == NeedsOne {
					response.New(http.StatusUnauthorized).Err("invalid role").Send(c)
					return
				}
			}
			// fmt.Println(userRoleTypes)
			c.Set("userRoleTypes", *userRoleTypes)

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

// retries creating a postgres account for a user, defaults to a level 0 user, aka, no special roles.
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

func createUserRoleTypes(roles []interface{}) (*UserRoleTypes, error) {
	userRoleTypes := UserRoleTypes{}
	pattern := `^mod_[0-9]+$`
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	for _, role := range roles {
		roleString := role.(string)
		switch roleString {
		case "admin":
			userRoleTypes.Admin = true
		case "mod":
			userRoleTypes.GlobalMod = true
		}
		if strings.Contains(roleString, "mod_") {
			res := regex.FindString(roleString)

			stringSchoolID := strings.Split(res, "_")[1]
			schoolID, err := strconv.Atoi(stringSchoolID)
			if err != nil {
				return nil, err
			}
			userRoleTypes.SchoolMods = append(userRoleTypes.SchoolMods, schoolID)
		}

	}

	return &userRoleTypes, nil
}
