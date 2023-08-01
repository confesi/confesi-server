package admin

import (
	"confesi/config"
	"confesi/db"
	"confesi/lib/logger"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

const (
	seenPostsCacheExpiry = 24 * time.Hour // one day
)

type AdminPostDetail struct {
	Post          db.Post `json:"post"`
	ReportCount   uint    `json:"-"`
	ReviewedByMod bool    `json:"-"`
}

func (h *handler) handleGetRankedPostsByReport(c *gin.Context) {
	// extract request
	var req validation.RankedPostsByReportsQuery
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
	postSpecificKey, err := utils.CreateCacheKey(config.RedisPostsCacheByReports, token.UID, req.SessionKey)
	if err != nil {
		response.New(http.StatusBadRequest).Err(utils.UuidError.Error()).Send(c)
		return
	}

	if req.PurgeCache {
		// purge the cache
		err := h.redis.Del(c, postSpecificKey).Err()
		if err != nil {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	// retrieve the seen post IDs from the cache
	ids, err := h.redis.SMembers(c, postSpecificKey).Result()
	if err != nil {
		if err == redis.Nil {
			ids = []string{} // assigns an empty slice
		} else {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	excludedIDQuery := ""
	if len(ids) > 0 {
		excludedIDQuery = " AND posts.id NOT IN (" + strings.Join(ids, ",") + ")"
	}

	posts := []db.Post{}
	// fetch comments
	err = h.db.
		Preload("Faculty").
		Preload("School").
		Preload("YearOfStudy").
		Where("reviewed_by_mod = ?"+excludedIDQuery, req.ReviewedByMod).
		Order("report_count DESC").
		Find(&posts).
		Limit(config.AdminPostsSortedByReportsPageSize).
		Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// to help with serialization
	adminPosts := []AdminPostDetail{}
	for i := range posts {
		err := h.redis.SAdd(c, postSpecificKey, posts[i].ID).Err()
		if err != nil {
			logger.StdErr(err, nil, nil, nil, nil)
			response.New(http.StatusInternalServerError).Err("failed to update cache").Send(c)
			return
		}
		post := &posts[i]
		if !utils.ProfanityEnabled(c) {
			*post = post.CensorPost()
		}
		// for every post, make it a post admin detail
		adminPosts = append(adminPosts, AdminPostDetail{Post: *post, ReportCount: post.ReportCount, ReviewedByMod: post.ReviewedByMod})
	}

	// set the expiration for the cache
	err = h.redis.Expire(c, postSpecificKey, seenCommentsCacheExpiry).Err()
	if err != nil {
		logger.StdErr(err, nil, nil, nil, nil)
		response.New(http.StatusInternalServerError).Err("failed to set cache expiration").Send(c)
		return
	}

	// Send response
	response.New(http.StatusOK).Val(adminPosts).Send(c)
}
