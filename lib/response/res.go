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

type apiResult struct {
	Code      int         `json:"-"`
	Error     interface{} `json:"error"`
	ErrorCode interface{} `json:"error_code"`
	Value     interface{} `json:"value"`
}

func (r *apiResult) Err(err string) *apiResult {
	r.Error = err
	return r
}

func (r *apiResult) Val(value interface{}) *apiResult {
	r.Value = value
	return r
}

func (r *apiResult) ErrCode(errCode int) *apiResult {
	r.ErrorCode = errCode
	return r
}

func New(code int) *apiResult {
	return &apiResult{
		Code: code,
	}
}

func (r *apiResult) Send(c *gin.Context) {
	if r.Error != nil {
		logger.ResErr(errors.New(fmt.Sprintf("%v", r.Error)), c, r.Code)
	}
	c.JSON(r.Code, r)
	c.Abort() // added back, because without it things break
}
