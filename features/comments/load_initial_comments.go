package comments

import (
	"confesi/config"
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

func fetchComments(postID int64, gm *gorm.DB, excludedIDs []string, sort string) ([]CommentDetail, error) {
	var comments []CommentDetail

	excludedIDQuery := ""
	if len(excludedIDs) > 0 {
		excludedIDQuery = "AND comments.id NOT IN (" + strings.Join(excludedIDs, ",") + ")"
	}

	var sortField string
	switch sort {
	case "new":
		sortField = "created_at DESC"
	case "trending":
		sortField = "score DESC" // todo: make trending_score eventually once voting is added to comments
	default:
		// should never happen with validated struct, but to be defensive
		logger.StdErr(errors.New(fmt.Sprintf("invalid sort type: %q", sort)))
		return nil, errors.New("invalid sort field")
	}
	query := gm.
		Preload("Identifiers").
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
				   ROW_NUMBER() OVER (PARTITION BY c.ancestors[1] ORDER BY c.created_at DESC) AS reply_num
			FROM comments c
			JOIN top_root_comments tr ON c.ancestors[1] = tr.id
		)
		SELECT t.id, t.post_id, t.score, t.content, t.ancestors, t.created_at, t.updated_at, t.hidden, t.children_count, t.user_id, t.downvote, t.upvote, t.user_vote
		FROM (
			SELECT combined_comments.id, combined_comments.post_id, combined_comments.score, combined_comments.content, combined_comments.ancestors, combined_comments.created_at, combined_comments.updated_at, combined_comments.hidden, combined_comments.children_count, combined_comments.user_id, combined_comments.downvote, combined_comments.upvote,
				   COALESCE(
					   (SELECT votes.vote
						FROM votes
						WHERE votes.comment_id = combined_comments.id
						  AND votes.user_id = combined_comments.user_id
						LIMIT 1),
					   '0'::vote_score_value
				   ) AS user_vote
			FROM (
				SELECT id, post_id, score, content, ancestors, updated_at, created_at, hidden, user_id, children_count, downvote, upvote FROM top_root_comments
				UNION ALL
				SELECT id, post_id, score, content, ancestors, created_at, updated_at, hidden, user_id, children_count, downvote, upvote
				FROM ranked_replies
				WHERE reply_num <= ?
			) AS combined_comments
		) AS t;
		
    `, postID, config.RootCommentsLoadedInitially, config.RepliesLoadedInitially).
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

	// session key ("comments" + ":" + "userid" + "+" + "[uuid_session]") -> root comment ids seen
	postSpecificKey := idSessionKey

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

		// if comment is hidden, set its content to "[removed]"
		if comment.Comment.Hidden {
			comment.Comment.Content = "[removed]"
		}

		err := h.redis.SAdd(c, postSpecificKey, fmt.Sprint(comment.Comment.ID)).Err()
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
