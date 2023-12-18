package public

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleIndex(c *gin.Context) {
	currentDir, err := os.Getwd()
	if err != nil {
		c.String(http.StatusInternalServerError, "Unknown error")
		return
	}

	htmlFilePath := filepath.Join(currentDir, "./public/html", "index.html")
	htmlContent, err := ioutil.ReadFile(htmlFilePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "Unknown error")
		return
	}

	c.Data(http.StatusOK, "text/html", htmlContent)
}
