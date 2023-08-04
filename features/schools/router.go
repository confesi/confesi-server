package schools

// Distance calculations are inputted in meters, and outputted in kilometers

import (
	"confesi/db"
	"confesi/lib/cache"
	"confesi/lib/fire"
	"confesi/middleware"
	"errors"
	"math"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

var (
	serverError = errors.New("server error")
	invalidId   = errors.New("invalid id")
)

type handler struct {
	*gorm.DB
	fb    *fire.FirebaseApp
	redis *redis.Client
}

// value that gets sent back to client for each of their watched schools
type SchoolDetail struct {
	db.School
	Home     bool     `json:"home"`
	Watched  bool     `json:"watched"`
	Distance *float64 `json:"distance"`
}

type Coordinate struct {
	lat    float64
	lon    float64
	radius float64
}

// todo: could use this in future: https://socketloop.com/tutorials/golang-find-location-by-ip-address-and-display-with-google-map
// Algo from:
// https://stackoverflow.com/a/365853
func (c *Coordinate) getDistance(dest db.School) float64 {
	const r float64 = 6371 // earth radius
	destLat := degreeToRad(float64(dest.Lat))
	originLat := degreeToRad(c.lat)

	deltaLat := degreeToRad(float64(dest.Lat) - c.lat)
	deltaLon := degreeToRad(float64(dest.Lon) - c.lon)

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Sin(deltaLon/2)*math.Sin(deltaLon/2)*math.Cos(destLat)*math.Cos(originLat)

	b := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return r * b // in kilometers
}

func Router(r *gin.RouterGroup) {
	h := handler{db.New(), fire.New(), cache.New()}

	// any firebase user
	anyFirebaseUser := r.Group("")
	anyFirebaseUser.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.AllFbUsers, []string{})
	})
	anyFirebaseUser.GET("/rank", h.handleGetRankedSchools)
	anyFirebaseUser.DELETE("/purge", h.handlePurgeRankedSchoolsCache)
	anyFirebaseUser.GET("/random", h.handleGetRandomSchool)
	anyFirebaseUser.GET("/", h.handleGetSchools)
	anyFirebaseUser.GET("/search", h.handleGetSchoolsByQuery)

	// only registered firebase users
	registeredFirebaseUsersOnly := r.Group("")
	registeredFirebaseUsersOnly.Use(func(c *gin.Context) {
		middleware.UsersOnly(c, h.fb.AuthClient, middleware.RegisteredFbUsers, []string{})
	})
	registeredFirebaseUsersOnly.POST("/watch", h.handleWatchSchool)
	registeredFirebaseUsersOnly.DELETE("/unwatch", h.handleUnwatchSchool)
	registeredFirebaseUsersOnly.GET("/watched", h.handleGetWatchedSchools)
}
