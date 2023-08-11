package posts

import (
	"confesi/config"
	tags "confesi/lib/emojis"
	"confesi/lib/encryption"
	"confesi/lib/logger"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
	"strings"

	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/gin-gonic/gin"
)

const (
	seenPostsCacheExpiry = 24 * time.Hour // one day
)

// `all_schools` takes precedence over `school_id`
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

	if req.SchoolId == "" && !req.AllSchools {
		response.New(http.StatusBadRequest).Err("school id or all specification must be provided").Send(c)
		return
	}

	var unmaskedSchoolId uint
	if req.SchoolId != "" && !req.AllSchools {
		unmaskedSchoolId, err = encryption.Unmask(req.SchoolId)
		if err != nil {
			response.New(http.StatusBadRequest).Err("invalid school id").Send(c)
			return
		}
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
	case "sentiment":
		sortField = "sentiment DESC"
	default:
		// should never happen with validated struct, but to be defensive
		logger.StdErr(errors.New(fmt.Sprintf("invalid sort type: %q", req.Sort)))
		response.New(http.StatusBadRequest).Err("invalid sort field").Send(c)
		return
	}

	var possibleExclusion string
	if len(ids) > 0 {
		cleanedIds := make([]string, len(ids))
		for i, id := range ids {
			cleanedIds[i] = strings.Trim(id, "{}") // remove curly braces
		}
		idsStr := strings.Join(cleanedIds, ", ") // convert the cleaned ids slice to a comma-separated string
		possibleExclusion = "posts.id NOT IN ( " + idsStr + " )"
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
		Preload("Category").
		Preload("YearOfStudy").
		Preload("Faculty").
		Order(sortField).
		Where(possibleExclusion).
		Where("hidden = ?", false)

	// if `all_schools` is true, then we don't need to filter by school
	if !req.AllSchools {
		query = query.Where("school_id = ?", unmaskedSchoolId)
	}

	err = query.
		Limit(config.FeedPostsPageSize).
		Find(&posts).Error
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
		if !utils.ProfanityEnabled(c) {
			post.Post = post.Post.CensorPost()
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
