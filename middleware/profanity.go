package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ----- Often used in combo with utils function like, ex:
// if !utils.ProfanityEnabled(c) {
// 	post.Post = post.Post.CensorPost()
// }
// -------------------------------------------------------

// Sets to context if profanity is allowed.
//
// Defaults to true unless specified.
func OptionalProfanityCensor(c *gin.Context) {
	profanityStr := c.Query("profanity")

	// If the "profanity" query parameter is not provided, or it's empty, set the default value to true.
	if profanityStr == "" {
		c.Set("profanity", true)
		c.Next()
		return
	}

	// Convert the "profanity" query parameter to a boolean value.
	profanity, err := strconv.ParseBool(profanityStr)
	if err != nil {
		// If there's an error parsing the boolean value, assume "profanity" is true and log the error.
		c.Set("profanity", true)
		c.Error(err)
		c.Next()
		return
	}

	c.Set("profanity", profanity)
	c.Next()
}
