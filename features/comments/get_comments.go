package comments

import (
	"confesi/db"
	"confesi/lib/logger"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

const (
	seenCommentsCacheExpiry = 24 * time.Hour // one day
)

func fetchComments(postID int64, limit int, gm *gorm.DB, excludedIDs []string) ([]db.Comment, error) {
	var comments []db.Comment

	excludedIDQuery := ""
	if len(excludedIDs) > 0 {
		excludedIDQuery = "AND comments.id NOT IN (" + strings.Join(excludedIDs, ",") + ")"
	}

	query := gm.
		Raw(`
			WITH RECURSIVE comment_hierarchy AS (
				SELECT *
				FROM comments
				WHERE ancestors[1] = ?
				AND comments.post_id = ?
				`+excludedIDQuery+`
				
				UNION
				
				SELECT c.*
				FROM comments c
				INNER JOIN comment_hierarchy ch ON ch.id = ANY(c.ancestors)
				WHERE ARRAY_LENGTH(c.ancestors, 1) = ARRAY_LENGTH(ch.ancestors, 1) + 1
			)
			SELECT *
			FROM comment_hierarchy
			ORDER BY score DESC
			LIMIT ?;
		`, 34, postID, limit).Find(&comments)

	if query.Error != nil {
		return nil, query.Error
	}

	return comments, nil
}

func (h *handler) handleGetComments(c *gin.Context) {
	// extract request
	var req validation.CommentQuery
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
	idSessionKey, err := utils.CreateCacheKey("comments", token.UID, req.SessionKey)
	if err != nil {
		response.New(http.StatusBadRequest).Err(utils.UuidError.Error()).Send(c)
		return
	}

	// session key (posts:userid+uuid_session) -> post id -> comment ids seen for that post
	postSpecificKey := idSessionKey + ":" + fmt.Sprint(req.PostID)

	if req.PurgeCache {
		// purge the cache
		err := h.redis.Del(c, postSpecificKey).Err()
		if err != nil {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	// retrieve the seen comment IDs from the cache
	ids, err := h.redis.SMembers(c, postSpecificKey).Result()
	if err != nil {
		if err == redis.Nil {
			ids = []string{} // assigns an empty slice
		} else {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	// fetch comments using the translated SQL query
	comments, err := fetchComments(int64(req.PostID), 3, h.db, ids)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// update the cache with the retrieved post IDs
	for _, comment := range comments {
		err := h.redis.SAdd(c, postSpecificKey, fmt.Sprint(comment.ID)).Err()
		if err != nil {
			logger.StdErr(err)
			response.New(http.StatusInternalServerError).Err("failed to update cache").Send(c)
			return
		}
	}

	// set the expiration for the cache
	err = h.redis.Expire(c, postSpecificKey, seenCommentsCacheExpiry).Err()
	if err != nil {
		logger.StdErr(err)
		response.New(http.StatusInternalServerError).Err("failed to set cache expiration").Send(c)
		return
	}

	// Send response
	response.New(http.StatusOK).Val(comments).Send(c)
}
