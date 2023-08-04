package schools

import (
	"confesi/config"
	"confesi/lib/logger"
	"confesi/lib/response"
	"confesi/lib/utils"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	limitQueryMax = 100

	latValueMax = 90
	latValueMin = -90

	longValueMax = 180
	longValueMin = -180
)

type Response struct {
	*Pagination
	Schools []SchoolDetail `json:"schools"`
}

type Pagination struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// NOTE: ignoring `lat` param and `long` param query if `school` is provided
func (h *handler) handleGetSchools(c *gin.Context) {
	pagination, err := getPagination(c)
	if err != nil {
		response.
			New(http.StatusBadRequest).
			Err(err.Error()).
			Send(c)
		return
	}

	schoolName := c.Query("school")
	radiusStr := c.Query("radius")

	latlong, err := utils.GetLatLong(c)
	if err != nil {
		logger.StdErr(err)
		response.
			New(http.StatusInternalServerError).
			Err(serverError.Error()).
			Send(c)
		return
	}
	lat := latlong.Lat
	long := latlong.Long

	missingLatLong := lat == 0 || long == 0
	if schoolName == "" && missingLatLong {
		response.
			New(http.StatusBadRequest).
			Err("not using location for schools list: no peer address").
			Send(c)
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	/* If `school` param is found */
	if schoolName != "" {
		var schools []SchoolDetail
		if err := h.getBySchoolName(&schools, schoolName, pagination, token.UID); err != nil {
			logger.StdErr(err)
			response.
				New(http.StatusInternalServerError).
				Err(err.Error()).
				Send(c)
			return
		}

		var schoolReponse []SchoolDetail
		for _, school := range schools {
			schoolReponse = append(schoolReponse, school)
		}

		response.
			New(http.StatusOK).
			Val(Response{pagination, schoolReponse}).
			Send(c)
		return
	}

	/* If `lat` and `long` is supplied */
	var schools []SchoolDetail
	if err := h.getAllSchools(&schools, token.UID); err != nil {
		logger.StdErr(err)
		response.
			New(http.StatusInternalServerError).
			Err(err.Error()).
			Send(c)
		return
	}

	// default value
	if radiusStr == "" {
		radiusStr = fmt.Sprintf("%d", config.DefaultRange)
	}

	coord, err := getCoord(lat, long, radiusStr)
	if err != nil {
		response.
			New(http.StatusBadRequest).
			Err(err.Error()).
			Send(c)
		return
	}

	var schoolsInRange []SchoolDetail // Use slice instead of an array

	for _, school := range schools {
		distance := coord.getDistance(school.School)
		if distance <= coord.radius {
			school.Distance = &distance
			schoolsInRange = append(schoolsInRange, school) // Append the school to the slice
		}
	}

	// Sort the schoolsInRange slice by "distance" in ascending order (smallest first)
	sort.Slice(schoolsInRange, func(i, j int) bool {
		return schoolsInRange[i].Distance != nil && schoolsInRange[j].Distance != nil && *schoolsInRange[i].Distance < *schoolsInRange[j].Distance
	})

	start := pagination.Offset
	if start > len(schoolsInRange) {
		start = 0
	}

	end := pagination.Offset + pagination.Limit
	if end > len(schoolsInRange) {
		end = len(schoolsInRange)
	}

	// Check if the schoolsInRange slice is empty and return an empty slice
	if len(schoolsInRange) == 0 {
		response.
			New(http.StatusOK).
			Val(Response{pagination, []SchoolDetail{}}).
			Send(c)
	} else {
		response.
			New(http.StatusOK).
			Val(Response{pagination, schoolsInRange[start:end]}).
			Send(c)
	}
}

func degreeToRad(deg float64) float64 {
	return (float64(deg) * (math.Pi / 180))
}

func getCoord(lat float64, long float64, radiusStr string) (*Coordinate, error) {

	if lat < float64(latValueMin) || lat > float64(latValueMax) {
		return nil, errors.New("lat value out of bound")
	}

	if long < float64(longValueMin) || long > float64(longValueMax) {
		return nil, errors.New("long value out of bound")
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		return nil, errors.New("invalid radius value")
	}

	return &Coordinate{lat, long, radius}, nil
}

func (h *handler) getAllSchools(schools *[]SchoolDetail, uid string) error {
	rawQuery := `
        SELECT schools.*, 
            COALESCE(u.school_id = schools.id, false) as home,
            CASE 
                WHEN EXISTS (SELECT 1 FROM school_follows WHERE user_id = ? AND school_id = schools.id)
                THEN true
                ELSE false
            END as watched
        FROM schools
        LEFT JOIN (
            SELECT DISTINCT school_id
            FROM users
            WHERE id = ?
        ) as u ON u.school_id = schools.id
    `

	err := h.DB.Raw(rawQuery, uid, uid).Scan(schools).Error
	return err
}

func (h *handler) getBySchoolName(
	schools *[]SchoolDetail,
	schoolName string,
	pag *Pagination,
	userID string,
) error {
	schoolSql := "%" + strings.ToUpper(schoolName) + "%"

	rawQuery := `
		SELECT schools.*, 
			COALESCE(u.school_id = schools.id, false) as home,
			CASE 
				WHEN EXISTS (SELECT 1 FROM school_follows WHERE user_id = ? AND school_id = schools.id)
				THEN true
				ELSE false
			END as watched
		FROM schools
		LEFT JOIN (
			SELECT DISTINCT school_id
			FROM users
			WHERE id = ?
		) as u ON u.school_id = schools.id
		WHERE name LIKE ? OR abbr LIKE ?
		OFFSET ? 
		LIMIT ?;
	`

	err := h.DB.Raw(rawQuery, userID, userID, schoolName, schoolSql, pag.Offset, pag.Limit).Scan(schools).Error
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

	if limit > limitQueryMax || limit < 0 {
		return nil, errors.New("limit query out of bound")
	}

	return &Pagination{
		Offset: int((offset - 1) * limit),
		Limit:  int(limit),
	}, nil
}
