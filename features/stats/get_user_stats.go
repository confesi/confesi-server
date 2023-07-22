package stats

import (
	"confesi/db"
	"confesi/lib/response"
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

	// token, err := utils.UserTokenFromContext(c)
	// if err != nil {
	// 	response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
	// 	return
	// }
	//! This is a temporary UID - will be replaced with the UID from the token
	UID := "KoO2S3suuPbYeIzwcN6ekYVIGtJ2"
	//Obtain the posts from the user
	posts := []db.Post{}

	err := h.db.Model(&db.Post{}).Where("user_id = ?", UID).Find(&posts).Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
	}
	// Initialize the user stats struct
	stats := UserStats{}

	//Obtain the stats from posts
	for _, post := range posts {
		stats.Likes += int(post.Upvote)
		stats.Dislikes += int(post.Downvote)
		if post.HottestOn != nil {
			stats.Hottest += 1
		}
	}

	response.New(http.StatusOK).Val(stats).Send(c)
}
