package schools

import (
	"confesi/db"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var h handler = handler{db.New()}

func routerSetup() *gin.Engine {
	mux := gin.Default()
	mux.GET("/schools", h.getSchools)
	return mux
}

func testSetup(path string) (*httptest.ResponseRecorder, *http.Request) {
	r, err := http.NewRequest("GET", path, nil)
	if err != nil {
		log.Fatal(err)
	}
	return httptest.NewRecorder(), r
}

func TestGetSchoolByName(t *testing.T) {
	var schools []db.School
	err := h.getBySchoolName(&schools, "uvic", &Pagination{Offset: 1, Limit: 20})
	assert.Nil(t, err)
}

func TestGetSchoolByCoord(t *testing.T) {
	mux := routerSetup()
	path := "/schools?offset=1&limit=10&lat=40.799391&lon=-77.860863&radius=800"
	w, r := testSetup(path)
	mux.Handler().ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
}

func TestGetSchoolBadRequest(t *testing.T) {
	testCases := []string{
		// missing offset
		"/schools?limit=10&lat=40.799391&lon=-77.860863&radius=800",
		// missing limit
		"/schools?offset=1&lat=40.799391&lon=-77.860863&radius=800",
		// missing lat
		"/schools?offset=1&limit=10&lon=-77.860863&radius=800",
		// missing lon
		"/schools?offset=1&limit=10&lat=40.799391&radius=800",
		// missing radius
		"/schools?offset=1&limit=10&lat=40.799391&lon=-76.860863",
		// either `school`, `lat`, `lon` are specified
		"/schools?offset=1&limit=10",
	}

	mux := routerSetup()
	for _, testCase := range testCases {
		w, r := testSetup(testCase)
		mux.ServeHTTP(w, r)
		log.Println(w.Code)
		//assert.Equal(t, w.Code, http.StatusBadRequest)
	}
}
