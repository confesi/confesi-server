package utils

import (
	"github.com/gin-gonic/gin"
)

type apiRes struct {
	Error      interface{} `json:"error"`
	Value      interface{} `json:"value"`
	StatusCode int         `json:"-"`
}

// Allows for a standardized response format across the api.
//
// Example usage: utils.Res(c, http.StatusOK, nil, gin.H{"email": "test@example.com"})
//
// Which returns json like:
//
//	{
//	 "error": null,
//	 "value": {
//	   "email": "test@example.com"
//	 }
//	}
func Res(c *gin.Context, statusCode int, err interface{}, value interface{}) {
	response := apiRes{
		Value:      value,
		Error:      err,
		StatusCode: statusCode,
	}

	c.JSON(statusCode, response)
	c.Abort()
}
