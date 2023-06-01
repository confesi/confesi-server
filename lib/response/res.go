package response

import (
	"confesi/lib"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

// Api result standardization. Any errors logged to StdErr. Usage:
//
// response.New(http.StatusAccepted).
//		Err("Some error message").
//		Val(gin.H{"name": "John"}).
//		Send(c)

type apiResult struct {
	Code  int         `json:"-"`
	Error interface{} `json:"error"`
	Value interface{} `json:"value"`
}

func (r *apiResult) Err(err string) *apiResult {
	r.Error = err
	return r
}

func (r *apiResult) Val(value interface{}) *apiResult {
	r.Value = value
	return r
}

func New(code int) *apiResult {
	return &apiResult{
		Code: code,
	}
}

func (r *apiResult) Send(c *gin.Context) {
	if r.Error != nil {
		errString := fmt.Sprintf("[status_code: %d], %v", r.Code, r.Error)
		lib.StdErr(errors.New(errString))
	}
	c.JSON(r.Code, r)
}
