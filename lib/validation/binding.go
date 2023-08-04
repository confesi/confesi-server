package validation

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

type DefaultBinding struct {
	Validator *validator.Validate
}

// Required to implement the Binding interface
func (b *DefaultBinding) Name() string {
	return "defaultBinding"
}

// Required to implement the Binding interface
//
// Accepts a pointer to a struct and binds the request body to it if it's valid.
func (b *DefaultBinding) Bind(req *http.Request, obj interface{}) error {
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(obj); err != nil {
		return err
	}

	if err := b.Validator.Struct(obj); err != nil {
		// handle validation errors
		validationErrs, ok := err.(validator.ValidationErrors)
		if ok {
			// handles specific validation errors
			// this returns the first error it encounters that's missing (i.e., "required")
			for _, validationErr := range validationErrs {
				fieldName := validationErr.Field()
				tag := validationErr.Tag()
				if tag == "required" {
					return &ValidationError{
						Field: fieldName,
						Tag:   tag,
					}
				}
			}
		}
		return err
	}

	return nil
}

// ValidationError is a custom error type that is returned when a validation error occurs.
//
// Implements the error interface.
type ValidationError struct {
	Field string
	Tag   string
}

func (v *ValidationError) Error() string {
	return v.Field + " validation failed for tag: " + v.Tag
}

// Custom validation function for "required_without" tag
func RequiredWithout(fl validator.FieldLevel) bool {
	field := fl.Field()
	otherFieldName := fl.Param()

	// get the other field value
	otherField := fl.Parent().FieldByName(otherFieldName)
	if !otherField.IsValid() {
		// aka, not found
		return false
	}
	otherFieldValue := strings.TrimSpace(otherField.String())

	// trim spaces from the current field value (ex: "  " shouldn't be considered valid)
	fieldValue := strings.TrimSpace(field.String())

	// if the other field is empty and the current field is also empty, return false
	if otherFieldValue == "" && fieldValue == "" {
		return false
	}

	return true
}
