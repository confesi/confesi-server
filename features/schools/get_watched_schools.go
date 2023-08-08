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
	var schools []SchoolDetail // Declare schools as a slice of SchoolDetail
	var userSchool *SchoolDetail

	// Retrieve schools with their home and watched status
	query := tx.Raw(`
		SELECT 
			s.*, 
			COALESCE(u.school_id = s.id, false) AS home,
			EXISTS (SELECT 1 FROM school_follows WHERE user_id = ? AND school_id = s.id) AS watched
		FROM schools AS s
		JOIN school_follows AS sf ON sf.school_id = s.id AND sf.user_id = ?
		JOIN users AS u ON u.id = ?
	`, token.UID, token.UID, token.UID)

	err := query.Find(&schools).Error
	if err != nil {
		return nil, serverError
	}

	if includeHomeSchool {
		// Fetch the user's school directly in the initial query using a join
		query := tx.Raw(`
			SELECT 
				schools.*, 
				EXISTS (SELECT 1 FROM school_follows WHERE user_id = ? AND school_id = schools.id) AS watched
			FROM schools 
			JOIN users ON users.school_id = schools.id
			WHERE users.id = ?
			AND schools.id = users.school_id
			LIMIT 1
		`, token.UID, token.UID)
		err := query.Find(&userSchool).Error
		if err != nil {
			return nil, serverError
		}
		userSchool.Home = true // if the user's school is included, it is always the home school
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
