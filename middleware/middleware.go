package middleware

import (
	"os"
)

var (
	token string
)

func init() {
	token = os.Getenv("APPCHECK_TOKEN")
	if token == "" {
		panic("`APPCHECK_TOKEN` env not set")
	}
}
