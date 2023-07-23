package stats

import (
	"confesi/config"
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type UserStats struct {
	Likes         int     `json:"likes"`
	Dislikes      int     `json:"dislikes"`
	Hottest       int     `json:"hottest"`
	Likes_Perc    float64 `json:"likes_perc"`
	Dislikes_Perc float64 `json:"dislikes_perc"`
	Hottest_Perc  float64 `json:"hottest_perc"`
}

func (h *handler) handleGetUserStats(c *gin.Context) {
	// extract request

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		//! UNCOMMENT THIS WHEN TOKENS ARE BACK UP
		// response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		// return
	}
	fmt.Println(token)
	//! REMOVE THIS HARDCODED UID
	UID := "KoO2S3suuPbYeIzwcN6ekYVIGtJ2"

	// query the database for the user stats
	query := h.db.Model(db.Post{}).
		Select("SUM(upvote) AS likes, SUM(downvote) AS dislikes, COUNT(hottest_on) AS hottest").
		Where("user_id = ?", UID)

	if query.Error != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
	}

	// Initialize the user stats struct and obtain the values from the query
	userStats := UserStats{}
	query.Scan(&userStats)
	// Obtain the global stats
	globalStats, err := GetGlobalStats(c, h.redis, h.db)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		return
	}

	// Calculate the percentages relative to the user
	userStats.Likes_Perc = float64(userStats.Likes) / float64(globalStats.Likes)
	userStats.Dislikes_Perc = float64(userStats.Dislikes) / float64(globalStats.Dislikes)
	userStats.Hottest_Perc = float64(userStats.Hottest) / float64(globalStats.Hottest)

	response.New(http.StatusOK).Val(userStats).Send(c)
}

type GlobalUserStats struct {
	Likes    int `json:"likes"`
	Dislikes int `json:"dislikes"`
	Hottest  int `json:"hottest"`
}

// GetGlobalStats returns the global stats of the entire app
func GetGlobalStats(c *gin.Context, redis_store *redis.Client, database *gorm.DB) (*GlobalUserStats, error) {
	store := redis_store
	idSessionKey := config.RedisGlobalUserStats
	ctx := c.Request.Context()
	stats := GlobalUserStats{}

	// query the database for the user stats
	jsonString, err := store.Get(ctx, idSessionKey).Result()

	// Check whether a cache exists or not
	if err == redis.Nil {
		//If no cache exists create one
		// query the database for the global stats
		query := database.Model(db.Post{}).
			Select("SUM(upvote) AS likes, SUM(downvote) AS dislikes, COUNT(hottest_on) AS hottest")

		if query.Error != nil {
			return nil, query.Error
		}

		// obtain the values from the query

		query.Scan(&stats)
		// Convert stats to string
		statsString, err := json.Marshal(stats)
		if err != nil {
			return nil, err
		}
		// Store the stats in the cache
		err = store.Set(ctx, idSessionKey, string(statsString), time.Hour*24).Err()
		if err != nil {
			return nil, err
		}

	} else if err != nil {
		return nil, err
	} else {
		// If cache exists, unmarshal it (convert it to a struct)
		err = json.Unmarshal([]byte(jsonString), &stats)
		if err != nil {
			return nil, err
		}
	}

	// Return the global stats
	return &stats, nil

}
