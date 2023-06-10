package schools

import (
	"confesi/db"
	"confesi/lib/response"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Offset  uint        `json:"offset"`
	Schools []db.School `json:"schools"`
}

func (h *handler) getSchools(c *gin.Context) {
	offsetStr := c.Query("offset")
	if strings.EqualFold(offsetStr, "") {
		offsetStr = "1"
	}

	offset, err := getOffset(offsetStr)
	if err != nil {
		response.
			New(http.StatusBadRequest).
			Err("invalid offset").
			Send(c)
		return

	}

	// by query, ignore lat and lon
	schoolName := c.Query("school")
	if schoolName != "" {
		response.
			New(http.StatusBadRequest).
			Err("invalid school query name").
			Send(c)
		return
	}

	// by lat and lon
	latStr := c.Query("lat")
	lonStr := c.Query("lon")
	if strings.EqualFold(latStr, "") || strings.EqualFold(lonStr, "") {
		response.
			New(http.StatusBadRequest).
			Err("missing coordinate information").
			Send(c)
		return
	}

	lat, err := validateCoord(latStr)
	if err != nil {
		response.
			New(http.StatusBadRequest).
			Err("invalid lat").
			Send(c)
		return
	}

	lon, err := validateCoord(lonStr)
	if err != nil {
		response.
			New(http.StatusBadRequest).
			Err("invalid lon").
			Send(c)
		return
	}

}

func validateCoord(c string) (float32, error) {
	coord, err := strconv.ParseFloat(c, 32)
	return float32(coord), err
}

func getOffset(c string) (uint, error) {
	coord, err := strconv.ParseFloat(c, 32)
	return uint(coord), err
}
