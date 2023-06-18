package posts

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/validation"

	"fmt"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const (
	postPageSize = 10
)

func (h *handler) handleGetPosts(c *gin.Context) {
	// extract request
	var req validation.PostQuery

	// bind to validator
	binding := &validation.DefaultBinding{
		Validator: validator.New(),
	}
	if err := binding.Bind(c.Request, &req); err != nil {
		response.New(http.StatusBadRequest).Err(fmt.Sprintf("failed validation: %v", err)).Send(c)
		return
	}

	fmt.Println("PRINT 1")

	if req.PurgeCache {
		// Purge the cache
		err := h.redis.HDel(c, "posts", req.SessionKey).Err()
		if err != nil {
			response.New(http.StatusInternalServerError).Err("server error").Send(c)
			return
		}
	}

	fmt.Println("PRINT 2")

	values, err := h.redis.HGet(c, "posts", req.SessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			values = "" // Assign an empty string
			// Alternatively, you can assign an empty list:
			// values = []string{}
		} else {
			fmt.Println("error: ", err)
			response.New(http.StatusInternalServerError).Err("server error").Send(c)
			return
		}
	}

	// Rest of the code...

	fmt.Println("PRINT 3")

	// Convert values to a string of post ids
	ids := strings.Split(values, ",")

	fmt.Println("PRINT 4")

	// Select all posts
	var posts []db.Post
	h.db.
		Preload("School").
		Preload("Faculty").
		Find(&posts).
		Order("vote_score DESC").
		Where("id NOT IN (?)", ids).
		Where("hidden = ?", false).
		Limit(postPageSize)

	// Update the cache with the retrieved post IDs
	var postIDs []string
	for _, post := range posts {
		postIDs = append(postIDs, fmt.Sprint(post.ID))
	}

	// Convert postIDs to a string slice
	postIDsString := strings.Join(postIDs, ",")

	// Save the post IDs to the cache
	err = h.redis.HSet(c, "posts", req.SessionKey, postIDsString).Err()
	if err != nil {
		fmt.Println("error: ", err)
		response.New(http.StatusInternalServerError).Err("failed to save post IDs to cache").Send(c)
		return
	}

	// Send response
	response.New(http.StatusOK).Val(posts).Send(c)
}
