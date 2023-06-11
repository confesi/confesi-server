package schools

import (
	"confesi/db"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type handler struct {
	*gorm.DB
}

func Router(r *gin.RouterGroup) {
	h := handler{db.New()}
	//
	r.GET("/", h.getSchools)
}
