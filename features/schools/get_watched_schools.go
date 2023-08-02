package schools

import (
	"confesi/config"
	"confesi/lib/response"
	"confesi/lib/utils"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

func (h *handler) getWatchedSchools(c *gin.Context, token *auth.Token) ([]SchoolDetail, error) {
	schools := []SchoolDetail{}
	err := h.DB.
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
			) as u ON u.school_id = s.id;
		`, token.UID, token.UID).
		Find(&schools).Error
	if err != nil {
		return nil, serverError
	}
	return schools, nil
}

// TODO: should this be limited to only N schools? Paginated? Or
// TODO: will this be cached locally so we'd want to get everything?
func (h *handler) handleGetWatchedSchools(c *gin.Context) {
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	schools, err := h.getWatchedSchools(c, token)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
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
	response.New(http.StatusOK).Val(schools).Send(c)
}
