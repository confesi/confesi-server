package schools

import (
	"confesi/db"
	"confesi/lib/response"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Response struct {
	*Pagination
	Schools []db.School `json:"schools"`
}

type Pagination struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

func (h *handler) getSchools(c *gin.Context) {

	pagination, err := getPagination(c)
	if err != nil {
		response.
			New(http.StatusBadRequest).
			Err(err.Error()).
			Send(c)

		return

	}

	schoolName := c.Query("school")
	latStr := c.Query("lat")
	lonStr := c.Query("lon")

	if schoolName != "" {
		var schools []db.School
		if err := h.getBySchoolName(&schools, schoolName, pagination); err != nil {
			log.Println(err)
			response.
				New(http.StatusInternalServerError).
				Err(err.Error()).
				Send(c)
			return

		}

		response.
			New(http.StatusOK).
			Val(Response{pagination, schools}).
			Send(c)
		return
	} else if latStr == "" || lonStr == "" {
		response.
			New(http.StatusBadRequest).
			Err("invalid lat and lon query").
			Send(c)
		return
	}

}

func (h *handler) getBySchoolName(
	schools *[]db.School,
	schoolName string,
	pag *Pagination,
) error {
	schoolSql := "%" + strings.ToUpper(schoolName) + "%"
	err := h.DB.
		Table(db.Schools).
		Where("name LIKE ? OR abbr LIKE ?", schoolSql, schoolSql).
		Offset(pag.Offset).
		Limit(pag.Limit).
		Scan(schools).
		Error

	return err
}

func getPagination(c *gin.Context) (*Pagination, error) {
	offset, err := strconv.ParseInt(c.Query("offset"), 10, 32)
	if err != nil {
		return nil, errors.New("invalid page offset query")
	}
	if offset <= 0 {
		return nil, errors.New("invalid page offset value, offset must be greater than 0")
	}

	limit, err := strconv.ParseInt(c.Query("limit"), 10, 32)
	if err != nil {
		return nil, errors.New("invalid page limit query")
	}

	return &Pagination{
		Offset: int((offset - 1) * limit),
		Limit:  int(limit),
	}, nil
}
