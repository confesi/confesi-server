package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func appCheckInit(r *http.Request) int {
	mux := gin.New()
	mux.Use(AppCheck)
	mux.GET("/test", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	return w.Result().StatusCode
}

func TestAppCheckNoToken(t *testing.T) {
	r, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		panic(err)
	}

	statusCode := appCheckInit(r)
	assert.Equal(t, http.StatusForbidden, statusCode)
}

func TestAppCheckOK(t *testing.T) {
	token := os.Getenv("APPCHECK_TOKEN")
	if token == "" {
		panic("token not found")
	}

	r, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		panic(err)
	}
	r.Header.Add("X-AppCheck-Token", token)

	statusCode := appCheckInit(r)
	assert.Equal(t, http.StatusOK, statusCode)
}

func TestAppCheckWrongToken(t *testing.T) {
	r, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		panic(err)
	}
	r.Header.Add("X-AppCheck-Token", "wrong-token")

	statusCode := appCheckInit(r)
	assert.Equal(t, http.StatusForbidden, statusCode)
}
