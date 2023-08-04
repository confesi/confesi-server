package schools

import (
	"confesi/config"
	"confesi/lib/response"
	"confesi/lib/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleGetSchoolsByQuery(c *gin.Context) {
	query := c.Query("query")

	if query == "" {
		response.New(http.StatusBadRequest).Err("needs query").Send(c)
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// Construct the SQL query to fetch search results
	sqlQuery := `
	SELECT s.*, 
	COALESCE(u.school_id = s.id, false) as home,
	CASE 
	  WHEN EXISTS (SELECT 1 FROM school_follows WHERE user_id = ? AND school_id = s.id)
	  THEN true
	  ELSE false
	END as watched,
	SIMILARITY(name, ?) AS name_match,
	SIMILARITY(abbr, ?) AS abbr_match
	FROM schools as s
	LEFT JOIN (
	SELECT DISTINCT school_id
	FROM users
	WHERE id = ?
	) as u ON u.school_id = s.id
	WHERE (SIMILARITY(name, ?) + SIMILARITY(abbr, ?)) > ?
	ORDER BY (SIMILARITY(name, ?) + SIMILARITY(abbr, ?)) DESC
	LIMIT ?;
	`

	// Execute the query with the search query as a parameter
	var schools []SchoolDetail
	if err := h.DB.Raw(sqlQuery, token.UID, query, query, token.UID, query, query, config.QueryForSchoolsBySearchFloorSimilarityMatchValue, query, query, config.QueryForSchoolsBySearchPageSize).Scan(&schools).Error; err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// loop through schools
	for i := range schools {
		schoolDetail := &schools[i]
		latlong, err := utils.GetLatLong(c)
		if err == nil {
			coord := Coordinate{lat: latlong.Lat, lon: latlong.Long, radius: config.DefaultRange}
			distance := coord.getDistance(schoolDetail.School)
			schoolDetail.Distance = &distance
		}
	}

	// Check if schools is nil or empty and replace it with an empty slice
	if schools == nil || len(schools) == 0 {
		schools = []SchoolDetail{}
	}

	// Send the search results as a response
	response.New(http.StatusOK).Val(schools).Send(c)
}
