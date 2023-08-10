package schools

import (
	"confesi/config"
	"confesi/lib/response"
	"confesi/lib/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleGetRandomSchool(c *gin.Context) {
	schoolDetail := SchoolDetail{}

	withoutSchoolId := c.Query("without-school")

	// parse without-school parameter
	parsedWithoutSchoolId, err := strconv.ParseInt(withoutSchoolId, 10, 64)
	if err != nil {
		parsedWithoutSchoolId = -1 // Set to a value that won't match any school ID
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	query := `
		SELECT s.*, 
			COALESCE(u.school_id = s.id, false) as home,
			CASE 
				WHEN EXISTS (SELECT 1 FROM school_follows WHERE user_id = ? AND school_id = s.id)
				THEN true
				ELSE false
			END as watched
		FROM schools as s
		LEFT JOIN (
			SELECT DISTINCT school_id
			FROM users
			WHERE id = ?
		) as u ON u.school_id = s.id
	`

	// modify the query if without-school parameter is provided
	if parsedWithoutSchoolId > 0 {
		query += "WHERE s.id != ?"
	}

	// complete the query
	query += " ORDER BY RANDOM() LIMIT 1;"

	// prepare arguments for the query
	args := []interface{}{token.UID, token.UID}
	if parsedWithoutSchoolId > 0 {
		args = append(args, parsedWithoutSchoolId)
	}

	// execute the query
	err = h.DB.Raw(query, args...).Scan(&schoolDetail).Error
	if err != nil {
		// Handle error
	}

	latlong, err := utils.GetLatLong(c)
	if err == nil {
		coord := Coordinate{lat: latlong.Lat, lon: latlong.Long, radius: config.DefaultRange}
		distance := coord.getDistance(schoolDetail.School)
		schoolDetail.Distance = &distance
	}

	response.New(http.StatusOK).Val(schoolDetail).Send(c)
}
