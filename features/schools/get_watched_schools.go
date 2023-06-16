package schools

import (
	"confesi/lib/response"
	"confesi/lib/utils"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

// value that gets sent back to client for each of their watchd schools
type schoolResult struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Abbr   string `json:"abbr"`
	Lat    string `json:"lat"`
	Lon    string `json:"lon"`
	Domain string `json:"domain"`
}

func (h *handler) getWatchedSchools(c *gin.Context, token *auth.Token) ([]schoolResult, error) {
	schools := []schoolResult{}
	err := h.DB.
		Table("school_follows").
		Select("schools.id as id, schools.name, schools.abbr, schools.lat, schools.lon, schools.domain").
		Joins("JOIN schools ON school_follows.school_id = schools.id").
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
	response.New(http.StatusOK).Val(schools).Send(c)
}
