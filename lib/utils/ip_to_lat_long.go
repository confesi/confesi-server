package utils

import (
	"confesi/config"
	"confesi/middleware"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ip2location/ip2location-go/v9"
)

var ipDb *ip2location.DB

// `init` function is keyword func that is executed on package load (app start)
func init() {
	// open and initialize IP2Location database
	var err error
	ipDb, err = ip2location.OpenDB("./assets/IP2LOCATION/IP2LOCATION-LITE-DB5.IPV6.BIN")

	if err != nil {
		// if error, panic
		panic(fmt.Sprintf("Error opening IP2Location database: %s", err.Error()))
	}
}

func GetLatLong(c *gin.Context) (*middleware.LatLongCoord, error) {

	// Create lat long struct
	latLong := middleware.LatLongCoord{}

	// Obtain client IP address
	ip := c.ClientIP()

	//! REMOVE THIS AFTER TESTING (DEV_REMOVAL)
	if strings.Contains(ip, "172") && config.Development {
		ip = "38.240.226.38"
	}

	// Get lat long from IP
	lat, err := ipDb.Get_latitude(ip)

	if err != nil {
		return nil, err
	}

	// Get long from IP
	long, err := ipDb.Get_longitude(ip)

	if err != nil {
		return nil, err
	}
	// Convert lat and long from float32 to float64
	latLong.Lat = float64(lat.Latitude)
	latLong.Long = float64(long.Longitude)

	return &latLong, nil
}
