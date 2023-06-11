package schools

import (
	"confesi/db"
	"confesi/lib/logger"
	"confesi/lib/response"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Response struct {
	*Pagination
	Schools []SchoolInfo `json:"schools"`
}

type SchoolInfo struct {
	db.School
	Distance *float64 `json:"distance"`
}

type Pagination struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type coordinate struct {
	lat    float64
	lon    float64
	radius float64
}

// NOTE: ignoring `lat` param and `lon` param query if `school` is provided
func (h *handler) getSchools(c *gin.Context) {
	pagination, err := getPagination(c)
	if err != nil {
		response.
			New(http.StatusBadRequest).
			Err(err.Error()).
			Send(c)
		return

	}

	schoolName := c.Query("school")
	latStr := c.Query("lat")
	lonStr := c.Query("lon")
	radiusStr := c.Query("radius")

	missingLatLon := latStr == "" || lonStr == ""
	if schoolName == "" && missingLatLon {
		response.
			New(http.StatusBadRequest).
			Err("not using location for schools list: no peer address").
			Send(c)
		return
	}

	if schoolName != "" {
		var schools []db.School
		if err := h.getBySchoolName(&schools, schoolName, pagination); err != nil {
			log.Println(err)
			response.
				New(http.StatusInternalServerError).
				Err(err.Error()).
				Send(c)
			return

		}

		var schoolReponse []SchoolInfo
		for _, school := range schools {
			schoolReponse = append(schoolReponse, SchoolInfo{school, nil})
		}

		response.
			New(http.StatusOK).
			Val(Response{pagination, schoolReponse}).
			Send(c)
		return
	}

	var schools []db.School
	if err := h.getAllSchools(&schools); err != nil {
		logger.StdErr(err)
		response.
			New(http.StatusInternalServerError).
			Err(err.Error()).
			Send(c)
		return
	}

	coord, err := getCoord(latStr, lonStr, radiusStr)
	if err != nil {
		response.
			New(http.StatusBadRequest).
			Err(err.Error()).
			Send(c)
		return
	}

	var schoolsInRange []SchoolInfo
	for _, school := range schools {
		distance := coord.getDistance(school)
		if distance <= coord.radius {
			schoolsInRange = append(schoolsInRange, SchoolInfo{school, &distance})
		}
	}

	schoolCount := len(schoolsInRange)

	startingOffset := pagination.Offset
	if startingOffset > schoolCount {
		startingOffset = 0
	}

	endingOffset := pagination.Offset + pagination.Limit
	if endingOffset > schoolCount {
		endingOffset = schoolCount
	}

	response.
		New(http.StatusOK).
		Val(schoolsInRange[startingOffset:endingOffset]).
		Send(c)
}

// Algo from:
// https://stackoverflow.com/a/365853
func (c *coordinate) getDistance(dest db.School) float64 {
	const r float64 = 6371 // earth radius
	destLat := degreeToRad(float64(dest.Lat))
	originLat := degreeToRad(c.lat)

	deltaLat := degreeToRad(float64(dest.Lat) - c.lat)
	deltaLon := degreeToRad(float64(dest.Lon) - c.lon)

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Sin(deltaLon/2)*math.Sin(deltaLon/2)*math.Cos(destLat)*math.Cos(originLat)

	b := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return r * b // in km
}

func degreeToRad(deg float64) float64 {
	return (float64(deg) * (math.Pi / 180))
}

func getCoord(latStr, lonStr, radiusStr string) (*coordinate, error) {
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return nil, errors.New("invalid lat value")
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return nil, errors.New("invalid lon value")
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		return nil, errors.New("invalid radius value")
	}

	return &coordinate{lat, lon, radius}, nil
}

func (h *handler) getAllSchools(schools *[]db.School) error {
	return h.Find(schools).Error
}

func (h *handler) getBySchoolName(
	schools *[]db.School,
	schoolName string,
	pag *Pagination,
) error {
	schoolSql := "%" + strings.ToUpper(schoolName) + "%"
	err := h.DB.
		Table(db.Schools).
		Where("name LIKE ? OR abbr LIKE ?", schoolSql, schoolSql).
		Offset(pag.Offset).
		Limit(pag.Limit).
		Scan(schools).
		Error

	return err
}

func getPagination(c *gin.Context) (*Pagination, error) {
	offset, err := strconv.ParseInt(c.Query("offset"), 10, 32)
	if err != nil {
		return nil, errors.New("invalid page offset query")
	}
	if offset <= 0 {
		return nil, errors.New("invalid page offset value, offset must be greater than 0")
	}

	limit, err := strconv.ParseInt(c.Query("limit"), 10, 32)
	if err != nil {
		return nil, errors.New("invalid page limit query")
	}

	return &Pagination{
		Offset: int((offset - 1) * limit),
		Limit:  int(limit),
	}, nil
}
