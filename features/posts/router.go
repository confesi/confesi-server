package posts

import (
	"confesi/db"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type handler struct {
	db *gorm.DB
}

func Router(mux *gin.RouterGroup) {
	h := handler{db: db.New()}

	mux.POST("/create", func(c *gin.Context) {
		h.handleCreate(c)
	})
}
