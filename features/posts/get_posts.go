package posts

import (
	"confesi/config"
	tags "confesi/lib/emojis"
	"confesi/lib/logger"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"

	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/gin-gonic/gin"
)

const (
	seenPostsCacheExpiry = 24 * time.Hour // one day
)

// GetPosts godoc
//
//	@Summary		Get Posts.
//	@Description	Fetch posts.
//	@Tags			Posts
//	@Accept			application/json
//	@Produce		application/json
//	@Security		BearerAuth
//	@Security		X-AppCheck-Token
//	@Param			Body	body		string							true	"The title and/or body of the post"	SchemaExample({\n    "sort": "new",\n    "school": 1,\n    "purge_cache": true,\n    "session_key": "6ba7b810-9dad-11d1-80b4-00c04fd430c8"\n})
//	@Success		201		{object}	docs.Created					"Created"
//	@Failure		400		{object}	docs.InvalidSortField			"Invalid Sort Field"
//	@Failure		5001	{object}	docs.FailedToSetCacheExpiration	"Failed To Set Cache Expiration"
//	@Failure		5002	{object}	docs.FailedToUpdateCache		"Failed To Update Cache"
//	@Failure		5003	{object}	docs.ServerError				"Server Error"
//
//	@Router			/posts/posts [get]
func (h *handler) handleGetPosts(c *gin.Context) {
	// extract request
	var req validation.PostQuery
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// session key that can only be created by *this* user, so it can't be guessed to manipulate others' feeds
	idSessionKey, err := utils.CreateCacheKey(config.RedisPostsCache, token.UID, req.SessionKey)
	if err != nil {
		response.New(http.StatusBadRequest).Err(utils.UuidError.Error()).Send(c)
		return
	}

	if req.PurgeCache {
		// purge the cache
		err := h.redis.Del(c, idSessionKey).Err()
		if err != nil {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	// retrieve the post IDs from the cache
	ids, err := h.redis.SMembers(c, idSessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			ids = []string{} // assigns an empty slice
		} else {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	var sortField string
	switch req.Sort {
	case "new":
		sortField = "created_at DESC"
	case "trending":
		sortField = "trending_score DESC"
	default:
		// should never happen with validated struct, but to be defensive
		logger.StdErr(errors.New(fmt.Sprintf("invalid sort type: %q", req.Sort)))
		response.New(http.StatusBadRequest).Err("invalid sort field").Send(c)
		return
	}

	// select all posts that are not in the retrieved post IDs
	var posts []PostDetail
	query := h.db.
		Select(`
		posts.*,
		COALESCE(
			(
				SELECT votes.vote
				FROM votes
				WHERE votes.post_id = posts.id
				AND votes.user_id = ?
				LIMIT 1
			),
			'0'::vote_score_value
		) AS user_vote
		`, token.UID).
		Preload("School").
		Preload("Faculty").
		Order(sortField).
		Limit(config.FeedPostsPageSize).
		Where("hidden = ?", false).
		Where("school_id = ?", req.School)

	if len(ids) > 0 {
		query = query.Where("posts.id NOT IN (?)", ids)
	}

	err = query.Find(&posts).Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// update the cache with the retrieved post IDs
	for i := range posts {

		post := &posts[i]

		// check if the post belongs to the user
		if post.UserID == token.UID {
			post.Owner = true
		}
		post.Emojis = tags.GetEmojis(&post.Post)

		id := fmt.Sprint(post.ID)
		err := h.redis.SAdd(c, idSessionKey, id).Err()
		if err != nil {
			logger.StdErr(err)
			response.New(http.StatusInternalServerError).Err("failed to update cache").Send(c)
			return
		}
	}

	// set the expiration for the cache
	err = h.redis.Expire(c, idSessionKey, seenPostsCacheExpiry).Err()
	if err != nil {
		logger.StdErr(err)
		response.New(http.StatusInternalServerError).Err("failed to set cache expiration").Send(c)
		return
	}

	// Send response
	response.New(http.StatusOK).Val(posts).Send(c)
}
