package schools

import (
	"confesi/config"
	"confesi/lib/response"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleGetSchoolsByQuery(c *gin.Context) {
	query := c.Query("query")

	if query == "" {
		response.New(http.StatusBadRequest).Err("need query").Send(c)
		return
	}

	// Construct the SQL query to fetch search results
	sqlQuery := `
		SELECT *,
			SIMILARITY(name, ?) AS name_match,
			SIMILARITY(abbr, ?) AS abbr_match
		FROM schools
		WHERE (SIMILARITY(name, ?) + SIMILARITY(abbr, ?)) > ?
		ORDER BY (SIMILARITY(name, ?) + SIMILARITY(abbr, ?)) DESC
		LIMIT ?;
	`

	// Execute the query with the search query as a parameter
	var schools []School
	if err := h.DB.Raw(sqlQuery, query, query, query, query, config.QueryForSchoolsBySearchFloorSimilarityMatchValue, query, query, config.QueryForSchoolsBySearchPageSize).Scan(&schools).Error; err != nil {
		fmt.Println(err)
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// Check if schools is nil or empty and replace it with an empty slice
	if schools == nil || len(schools) == 0 {
		schools = []School{}
	}

	// Send the search results as a response
	response.New(http.StatusOK).Val(schools).Send(c)
}
