package drafts

import (
	"confesi/config"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FetchResults struct {
	Drafts []DraftDetail `json:"drafts"`
	Next   *int64        `json:"next"`
}

func (h *handler) handleGetYourDrafts(c *gin.Context) {
	// extract request
	var req validation.YourDraftsQuery
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	fetchResults := FetchResults{}

	err = h.db.
		Where("user_id = ?", token.UID).
		Order("updated_at DESC").
		Find(&fetchResults.Drafts).
		Where(req.Next.Cursor("updated_at >")).
		Limit(config.YourDraftsPageSize).
		Error

	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	if len(fetchResults.Drafts) > 0 {
		timeMicros := (fetchResults.Drafts[len(fetchResults.Drafts)-1].CreatedAt.Time).UnixMicro()
		fetchResults.Next = &timeMicros
	}

	response.New(http.StatusOK).Val(fetchResults).Send(c)
}
