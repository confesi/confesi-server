package public

import (
	"errors"
	"log"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

var (
	serverError = errors.New("server error")
)

type handler struct{}

func Router(mux *gin.RouterGroup) {
	h := handler{}

	// resolve the absolute path
	absPath, err := filepath.Abs("./public")
	if err != nil {
		log.Fatal(err)
	}

	// static files
	mux.Static("/public", absPath)

	// index
	mux.GET("/", h.handleIndex)
}
