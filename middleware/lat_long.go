package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type LatLongCoord struct {
	Lat  float64
	Long float64
}

func LatLong(c *gin.Context) {
	lat := c.Query("lat")
	long := c.Query("long")

	if lat == "" || long == "" {
		c.Set("latlong", nil)
		c.Next()
		return
	} else {
		latFloat, err := strconv.ParseFloat(lat, 64)
		if err != nil {
			c.Set("latlong", nil)
			c.Next()
			return
		}

		longFloat, err := strconv.ParseFloat(long, 64)
		if err != nil {
			c.Set("latlong", nil)
			c.Next()
			return
		}

		c.Set("latlong", LatLongCoord{Lat: latFloat, Long: longFloat})
		c.Next()
		return
	}
}
