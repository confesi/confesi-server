package stats

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserStats struct {
	Likes    int `json:"likes"`
	Dislikes int `json:"dislikes"`
	Hottest  int `json:"hottest"`
}

func (h *handler) handleGetUserStats(c *gin.Context) {
	// extract request

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		//! UNCOMMENT THIS WHEN TOKENS ARE BACK UP
		// response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		// return
	}
	fmt.Println(token)
	//! REMOVE THIS HARDCODED UID
	UID := "KoO2S3suuPbYeIzwcN6ekYVIGtJ2"

	// err = h.db.Model(&db.Post{}).Where("user_id = ?", token.UID).Find(&posts).Count.Error
	query := h.db.Model(db.Post{}).
		Select("SUM(upvote) AS likes, SUM(downvote) AS dislikes, COUNT(hottest_on) AS hottest").
		Where("user_id = ?", UID)
	if query.Error != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
	}
	// Initialize the user stats struct and obtain the values from the query
	stats := UserStats{}
	query.Scan(&stats)

	response.New(http.StatusOK).Val(stats).Send(c)
}
