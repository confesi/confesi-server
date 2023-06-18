package posts

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/validation"

	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const (
	postPageSize    = 5
	cacheExpiration = 1 * time.Minute
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

	if req.PurgeCache {
		// purge the cache
		err := h.redis.Del(c, "posts:"+req.SessionKey).Err()
		if err != nil {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	// retrieve the post IDs from the cache
	ids, err := h.redis.SMembers(c, "posts:"+req.SessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			ids = []string{} // assigns an empty slice
		} else {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	// select all posts that are not in the retrieved post IDs
	var posts []db.Post
	query := h.db.
		Preload("School").
		Preload("Faculty").
		Order("vote_score DESC").
		Limit(postPageSize).
		Where("hidden = ?", false)

	if len(ids) > 0 {
		query = query.Where("posts.id NOT IN (?)", ids)
	}

	err = query.Find(&posts).Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// update the cache with the retrieved post IDs
	for _, post := range posts {
		id := fmt.Sprint(post.ID)
		err := h.redis.SAdd(c, "posts:"+req.SessionKey, id).Err()
		if err != nil {
			fmt.Println("error: ", err)
			response.New(http.StatusInternalServerError).Err("failed to update cache").Send(c)
			return
		}
	}

	// set the expiration for the cache
	err = h.redis.Expire(c, "posts:"+req.SessionKey, cacheExpiration).Err()
	if err != nil {
		fmt.Println("error: ", err)
		response.New(http.StatusInternalServerError).Err("failed to set cache expiration").Send(c)
		return
	}

	// Send response
	response.New(http.StatusOK).Val(posts).Send(c)
}
