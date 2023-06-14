package votes

import (
	"confesi/db"
	"confesi/lib/fire"
	"confesi/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type handler struct {
	db *gorm.DB
	fb *fire.FirebaseApp
}

func Router(mux *gin.RouterGroup) {
	h := handler{db: db.New(), fb: fire.New()}

	// only allow registered users to vote
	mux.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers)
	})

	mux.PUT("/vote", h.handleVote)
}
