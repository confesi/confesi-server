package utils

import (
	"github.com/gin-gonic/gin"
)

func ProfanityEnabled(c *gin.Context) bool {
	// retrieve the "profanity" value from the context and perform a type assertion to get the boolean value.
	profanity, ok := c.Get("profanity")
	if !ok {
		// the "profanity" value is not set in the context, so it defaults to true.
		return true
	}

	// perform a type assertion to get the boolean value correctly.
	enabled, ok := profanity.(bool)
	if !ok {
		// the type assertion failed, so assume "profanity" is true.
		return true
	}

	return enabled
}
