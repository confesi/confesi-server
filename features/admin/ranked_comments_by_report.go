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
	seenCommentsCacheExpiry = 24 * time.Hour // one day
)

type AdminCommentDetail struct {
	Comment       db.Comment `json:"comment"`
	ReportCount   uint       `json:"-"`
	ReviewedByMod bool       `json:"-"`
}

func (h *handler) handleGetRankedCommentsByReport(c *gin.Context) {
	// extract request
	var req validation.RankedCommentsByReportsQuery
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
	commentSpecificKey, err := utils.CreateCacheKey(config.RedisCommentsCacheByReports, token.UID, req.SessionKey)
	if err != nil {
		response.New(http.StatusBadRequest).Err(utils.UuidError.Error()).Send(c)
		return
	}

	if req.PurgeCache {
		// purge the cache
		err := h.redis.Del(c, commentSpecificKey).Err()
		if err != nil {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	// retrieve the seen comment IDs from the cache
	ids, err := h.redis.SMembers(c, commentSpecificKey).Result()
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
		excludedIDQuery = " AND comments.id NOT IN (" + strings.Join(ids, ",") + ")"
	}

	comments := []db.Comment{}
	// fetch comments
	err = h.db.
		Where("reviewed_by_mod = ?"+excludedIDQuery, req.ReviewedByMod).
		Order("report_count DESC").
		Find(&comments).
		Limit(config.AdminCommentsSortedByReportsPageSize).
		Error
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// to help with serialization
	adminComments := []AdminCommentDetail{}
	for i := range comments {
		err := h.redis.SAdd(c, commentSpecificKey, comments[i].ID).Err()
		if err != nil {
			logger.StdErr(err, nil, nil, nil, nil)
			response.New(http.StatusInternalServerError).Err("failed to update cache").Send(c)
			return
		}
		comment := &comments[i]
		if !utils.ProfanityEnabled(c) {
			*comment = comment.CensorComment()
		}
		// for every comment, make it a comment admin detail
		adminComments = append(adminComments, AdminCommentDetail{Comment: *comment, ReportCount: comment.ReportCount, ReviewedByMod: comment.ReviewedByMod})
	}

	// set the expiration for the cache
	err = h.redis.Expire(c, commentSpecificKey, seenCommentsCacheExpiry).Err()
	if err != nil {
		logger.StdErr(err, nil, nil, nil, nil)
		response.New(http.StatusInternalServerError).Err("failed to set cache expiration").Send(c)
		return
	}

	// Send response
	response.New(http.StatusOK).Val(adminComments).Send(c)
}
