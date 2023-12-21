package user

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleGetAwards(c *gin.Context) {

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// Fetch awards the user already has
	userAwards := []db.AwardsTotal{}
	query := h.db.
		Preload("AwardType").
		Model(db.AwardsTotal{}).
		Where("user_id = ?", token.UID). // token.UID
		Find(&userAwards).
		Error
	if query != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// Fetch all possible awards
	allAwards := []db.AwardType{}
	query = h.db.Find(&allAwards).Error
	if query != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// Filter out awards the user already has
	missingAwards := []db.AwardType{}
	for _, award := range allAwards {
		hasAward := false
		for _, userAward := range userAwards {
			if userAward.AwardType.ID == award.ID {
				hasAward = true
				break
			}
		}
		if !hasAward {
			missingAwards = append(missingAwards, award)
		}
	}

	// Combine results
	result := struct {
		UserAwards    []db.AwardsTotal `json:"has"`
		MissingAwards []db.AwardType   `json:"missing"`
	}{
		UserAwards:    userAwards,
		MissingAwards: missingAwards,
	}

	response.New(http.StatusOK).Val(result).Send(c)
}
