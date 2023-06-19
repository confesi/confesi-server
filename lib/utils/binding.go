package utils

import (
	"confesi/lib/response"
	"confesi/lib/validation"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func Bind(c *gin.Context, req interface{}) {
	// validator
	binding := &validation.DefaultBinding{
		Validator: validator.New(),
	}
	if err := binding.Bind(c.Request, &req); err != nil {
		response.New(http.StatusBadRequest).Err(fmt.Sprintf("failed validation: %v", err)).Send(c)
		return
	}
}
