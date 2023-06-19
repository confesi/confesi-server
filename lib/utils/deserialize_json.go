package utils

import (
	"confesi/lib/response"
	"confesi/lib/validation"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// EXAMPLE USAGE

// var req validation.CreatePostDetails
// err := utils.New(c).ForceCustomTag("required_without", validation.RequiredWithout).Validate(&req)
// if err != nil {
// 	return
// }

// or

// var req validation.CreatePostDetails
// err := utils.New(c).Validate(&req)
// if err != nil {
// 	return
// }

// etc.

// EXAMPLE USAGE

type binding struct {
	validator *validator.Validate
	context   *gin.Context
}

// Adds a custom tag to the validator. Returns errors if tag is invalid.
func (b *binding) CustomTag(tag string, fn validator.Func) (*binding, error) {
	// registers a custom tag
	err := b.validator.RegisterValidation(tag, fn)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Adds a custom tag to the validator. Assumes the tag is valid for brevity in handlers.
// If the tag isn't valid, it panics.
func (b *binding) ForceCustomTag(tag string, fn validator.Func) *binding {
	// registers a custom tag
	b.validator.RegisterValidation(tag, fn)
	return b
}

// Creates a new validator instance.
func New(c *gin.Context) *binding {
	// create default binding struct
	b := &binding{}
	b.validator = validator.New()
	b.context = c
	return b
}

// Validates requeset json and deserializes it into the provided struct.
//
// Json must be a pointer to a struct.
func (b *binding) Validate(json interface{}) error {
	// ensure json is a pointer
	if reflect.ValueOf(json).Kind() != reflect.Ptr {
		response.New(http.StatusInternalServerError).Err("json must be referenced via pointer").Send(b.context)
		return errors.New("json must be referenced via pointer")
	}

	// validator
	defaultBinding := &validation.DefaultBinding{
		Validator: b.validator,
	}

	// check if valid
	if err := defaultBinding.Bind(b.context.Request, json); err != nil {
		response.New(http.StatusBadRequest).Err(fmt.Sprintf("failed validation: %v", err)).Send(b.context)
		return err
	}
	return nil
}
