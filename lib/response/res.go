package response

import (
	"confesi/lib/logger"
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

type ApiResult struct {
	Code  int         `json:"-"`
	Error interface{} `json:"error"`
	Value interface{} `json:"value"`
}

func (r *ApiResult) Err(err string) *ApiResult {
	r.Error = err
	return r
}

func (r *ApiResult) Val(value interface{}) *ApiResult {
	r.Value = value
	return r
}

func New(code int) *ApiResult {
	return &ApiResult{
		Code: code,
	}
}

func (r *ApiResult) Send(c *gin.Context) {
	if r.Error != nil {
		errString := fmt.Sprintf("[status_code: %d], %v", r.Code, r.Error)
		logger.StdErr(errors.New(errString))
	}
	c.JSON(r.Code, r)
	c.Abort() // added back, because without it things break
}
