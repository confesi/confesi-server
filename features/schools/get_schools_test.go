package schools

import (
	"confesi/db"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSchoolByName(t *testing.T) {
	h := handler{db.New()}
	var schools []db.School
	err := h.getBySchoolName(&schools, "uvic", &Pagination{Offset: 1, Limit: 20})
	assert.Nil(t, err)
}
