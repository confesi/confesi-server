package info

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"github.com/gin-gonic/gin"
)

func (h *handler) handleIndex(c *gin.Context) {
	// Get the absolute path of the running directory.
	currentDir, err := os.Getwd()
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting current directory")
		return
	}

	htmlFilePath := filepath.Join(currentDir, "pages", "index.html")

	htmlContent, err := ioutil.ReadFile(htmlFilePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error loading the HTML file")
		return
	}

	// Write the HTML content to the response with a 200 status code.
	c.Data(http.StatusOK, "text/html", htmlContent)
}
