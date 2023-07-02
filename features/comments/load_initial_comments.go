package comments

import (
	"confesi/config"
	"confesi/db"
	"confesi/lib/logger"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
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

func fetchComments(postID int64, gm *gorm.DB, excludedIDs []string, sort string) ([]db.Comment, error) {
	var comments []db.Comment

	excludedIDQuery := ""
	if len(excludedIDs) > 0 {
		excludedIDQuery = "AND comments.id NOT IN (" + strings.Join(excludedIDs, ",") + ")"
	}

	var sortField string
	switch sort {
	case "new":
		sortField = "created_at DESC"
	case "trending":
		sortField = "score DESC" // todo: make trending_score
	default:
		// should never happen with validated struct, but to be defensive
		logger.StdErr(errors.New(fmt.Sprintf("invalid sort type: %q", sort)))
		return nil, errors.New("invalid sort field")
	}
	query := gm.
		Raw(`
        WITH top_root_comments AS (
            SELECT *
            FROM comments
            WHERE COALESCE(ancestors, '{}') = '{}' AND post_id = ?
            `+excludedIDQuery+`
            ORDER BY `+sortField+`
            LIMIT ?
        ), ranked_replies AS (
            SELECT c.id, c.post_id, c.score, c.content, c.ancestors, c.created_at, c.updated_at, c.hidden, c.children_count, c.user_id, c.downvote, c.upvote,
            ROW_NUMBER() OVER (PARTITION BY c.ancestors[1] ORDER BY c.created_at) AS reply_num
            FROM comments c
            JOIN top_root_comments tr ON c.ancestors[1] = tr.id
        )
        SELECT id, post_id, score, content, ancestors, created_at, updated_at, hidden, children_count, user_id, downvote, upvote
        FROM (
            SELECT id, post_id, score, content, ancestors, updated_at, created_at, hidden, user_id, children_count, downvote, upvote FROM top_root_comments
            UNION ALL
            SELECT id, post_id, score, content, ancestors, created_at, updated_at, hidden, user_id, children_count, downvote, upvote
            FROM ranked_replies
            WHERE reply_num <= ?
        ) AS combined_comments
        ORDER BY (CASE WHEN cardinality(ancestors) = 0 THEN score END) DESC,
                 (CASE WHEN cardinality(ancestors) > 0 THEN created_at END) ASC;
    `, postID, config.RootsReturnedAtOnce, config.RepliesReturnedAtOnce).
		Find(&comments)

	if query.Error != nil {
		return nil, query.Error
	}

	return comments, nil
}

func (h *handler) handleGetComments(c *gin.Context) {
	// extract request
	var req validation.InitialCommentQuery
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

	// session key (posts:userid+uuid_session) -> post id -> root comment ids seen for that post
	postSpecificKey := idSessionKey + ":" + fmt.Sprint(req.PostID)

	if req.PurgeCache {
		// purge the cache
		err := h.redis.Del(c, postSpecificKey).Err()
		if err != nil {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	// retrieve the seen root comment IDs from the cache
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
	comments, err := fetchComments(int64(req.PostID), h.db, ids, req.Sort)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// update the cache with the retrieved post IDs
	for i := range comments {
		// access the comment using the index i (so I can change it
		// because loops are pass by value not reference)
		comment := &comments[i]

		// Rest of the code inside the loop
		fmt.Println(comment.Hidden)
		if comment.Hidden {
			comment.Content = "[removed]"
		}
		if len(comment.Ancestors) != 0 {
			// don't cache replies since they can be paginated
			// via next cursors on static created_at times, and subsequent
			// load_initial_comments will only fetch replies from uncached root comments
			continue
		}
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
