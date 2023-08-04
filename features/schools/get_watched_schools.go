package schools

import (
	"confesi/config"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type watchedResponse struct {
	Schools    []SchoolDetail `json:"schools"`
	UserSchool *SchoolDetail  `json:"user_school"`
}

func (h *handler) getWatchedSchools(c *gin.Context, token *auth.Token, tx *gorm.DB, includeHomeSchool bool) (*watchedResponse, error) {

	schools := []SchoolDetail{}
	var userSchool *SchoolDetail

	query := tx.
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
            ) as u ON true;
        `, token.UID, token.UID)

	if includeHomeSchool {
		// fetch the user's school directly in the initial query using a join
		query = query.Joins("JOIN users ON users.school_id = schools.id AND users.id = ?", token.UID)
	}

	// fetch user's school directly into the userSchool variable
	err := query.Scan(&schools).Error
	if err != nil {
		return nil, serverError
	}

	if !includeHomeSchool {
		userSchool = nil
	} else {
		// find the user's school in the list of schools
		for _, school := range schools {
			if school.Home {
				userSchool = &school
				break
			}
		}
	}

	return &watchedResponse{Schools: schools, UserSchool: userSchool}, nil
}

// TODO: should this be limited to only N schools? Paginated? Or
// TODO: will this be cached locally so we'd want to get everything?
func (h *handler) handleGetWatchedSchools(c *gin.Context) {

	// extract request
	var req validation.WatchedSchoolQuery

	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// start a transaction
	tx := h.DB.Begin()

	// if something goes ary, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}()

	res, err := h.getWatchedSchools(c, token, tx, req.IncludeHomeSchool)
	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		return
	}

	latlong, err := utils.GetLatLong(c)
	// loop through schools
	for i := range res.Schools {
		schoolDetail := &res.Schools[i]
		if err == nil {
			coord := Coordinate{lat: latlong.Lat, lon: latlong.Long, radius: config.DefaultRange}
			distance := coord.getDistance(schoolDetail.School)
			schoolDetail.Distance = &distance
		}
	}
	// for user school
	if req.IncludeHomeSchool {
		if err == nil {
			coord := Coordinate{lat: latlong.Lat, lon: latlong.Long, radius: config.DefaultRange}
			distance := coord.getDistance(res.UserSchool.School)
			res.UserSchool.Distance = &distance
		}
	}
	response.New(http.StatusOK).Val(res).Send(c)
}
