package votes

import (
	"confesi/db"
	"confesi/lib/fire"
	"confesi/middleware"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	serverError  = errors.New("server error")
	invalidValue = errors.New("invalid value")
)

type handler struct {
	db *gorm.DB
	fb *fire.FirebaseApp
	
}

func Router(mux *gin.RouterGroup) {
	h := handler{db: db.New(), fb: fire.New()}

	// only allow registered users to vote
	mux.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers, []string{})
	})

	mux.PUT("/vote", h.handleVote)
}
