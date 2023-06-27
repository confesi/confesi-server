package comments

import (
	"confesi/db"
	"confesi/lib/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleGetComments(c *gin.Context) {
	// Define the comments variable to store the query results
	var comments []db.Comment

	// Execute the query
	err := h.db.
		Where("ARRAY_LENGTH(ancestors, 1) = 1").
		Where("ancestors[1] = ?", 34).
		Order("updated_at").
		Limit(3).
		Find(&comments).
		Error

	if err != nil {
		// Handle the error
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// if all good, return 200 with the comments
	response.New(http.StatusOK).Val(comments).Send(c)
}
