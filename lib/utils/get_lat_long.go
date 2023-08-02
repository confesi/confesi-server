package utils

import (
	"confesi/middleware"
	"errors"

	"github.com/gin-gonic/gin"
)

func GetLatLong(c *gin.Context) (middleware.LatLongCoord, error) {
	// get latlong from context
	latLong, ok := c.Get("latlong")
	if !ok {
		return middleware.LatLongCoord{}, nil
	}
	latLongType, ok := latLong.(middleware.LatLongCoord)
	if !ok {
		return middleware.LatLongCoord{}, errors.New("bad type cast")
	}
	return latLongType, nil
}
