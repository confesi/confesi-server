package validation

// Custom validation tags

import (
	"confesi/db"

	"github.com/go-playground/validator/v10"
)

// Validates the year of study.
func YearOfStudyTag(v validator.FieldLevel) bool {
	value := v.Field().String()
	return value == db.YearOfStudyOne || value == db.YearOfStudyTwo || value == db.YearOfStudyThree ||
		value == db.YearOfStudyFour || value == db.YearOfStudyAlumniGraduate || value == db.YearOfStudyHidden
}
