package schools

import (
	"confesi/config"
	"confesi/lib/response"
	"confesi/lib/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleGetRandomSchool(c *gin.Context) {

	schoolDetail := SchoolDetail{}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	err = h.DB.
		Raw(`
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
			ORDER BY RANDOM() LIMIT 1;
		`, token.UID, token.UID).
		Scan(&schoolDetail).Error

	latlong, err := utils.GetLatLong(c)
	if err == nil {
		coord := Coordinate{lat: latlong.Lat, lon: latlong.Long, radius: config.DefaultRange}
		distance := coord.getDistance(schoolDetail.School)
		schoolDetail.Distance = &distance
	}

	response.New(http.StatusOK).Val(schoolDetail).Send(c)
}
