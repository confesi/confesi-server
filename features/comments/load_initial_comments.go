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

type CommentThreadGroup struct {
	Root    CommentDetail   `json:"root"`
	Replies []CommentDetail `json:"replies"`
	Next    *int64          `json:"next"`
}

const (
	seenCommentsCacheExpiry = 24 * time.Hour // one day

)

func fetchComments(postID int64, gm *gorm.DB, excludedIDs []string, sort string, uid string, h handler, c *gin.Context, commentSpecificKey string) ([]CommentThreadGroup, error) {
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
		sortField = "trending_score DESC"
	default:
		// should never happen with validated struct, but to be defensive
		logger.StdErr(errors.New(fmt.Sprintf("invalid sort type: %q", sort)))
		return nil, errors.New("invalid sort field")
	}
	// query written in raw SQL over pure Gorm because... well this would be a nightmare otherwise and likely impossible
	query := gm.
		Raw(`
		WITH top_root_comments AS (
			SELECT *
			FROM comments
			WHERE parent_root IS NULL AND post_id = ?
            `+excludedIDQuery+`
            ORDER BY `+sortField+`
			LIMIT ?
		), ranked_replies AS (
			SELECT c.id, c.post_id, c.vote_score, c.edited, c.trending_score, c.content, c.parent_root,
				CASE WHEN c.parent_root IS NOT NULL THEN c.created_at ELSE tr.created_at END AS created_at,
				CASE WHEN c.parent_root IS NOT NULL THEN c.updated_at ELSE tr.updated_at END AS updated_at,
				c.hidden, c.children_count, c.user_id, c.downvote, c.upvote, c.numerical_user, c.numerical_replying_user,
				c.numerical_replying_user_is_op, c.numerical_user_is_op,
				ROW_NUMBER() OVER (PARTITION BY c.parent_root ORDER BY c.created_at ASC) AS reply_num
			FROM comments c
			JOIN top_root_comments tr ON c.parent_root = tr.id
			ORDER BY c.created_at ASC
		)
		SELECT t.id, t.post_id, t.vote_score, t.edited, t.trending_score, t.content, t.parent_root,
			t.created_at, t.updated_at, t.hidden, t.children_count, t.user_id, t.downvote, t.upvote,
			t.numerical_user, t.numerical_replying_user, t.numerical_replying_user_is_op, t.numerical_user_is_op, t.user_vote
		FROM (
			SELECT combined_comments.id, combined_comments.post_id, combined_comments.vote_score, combined_comments.edited,
				combined_comments.trending_score, combined_comments.content, combined_comments.parent_root,
				combined_comments.created_at, combined_comments.updated_at, combined_comments.hidden,
				combined_comments.children_count, combined_comments.user_id, combined_comments.downvote,
				combined_comments.upvote, combined_comments.numerical_user, combined_comments.numerical_replying_user,
				combined_comments.numerical_replying_user_is_op, combined_comments.numerical_user_is_op,
				COALESCE(
					(SELECT votes.vote
					FROM votes
					WHERE votes.comment_id = combined_comments.id
					AND votes.user_id = ?
					LIMIT 1),
					'0'::vote_score_value
				) AS user_vote
			FROM (
				SELECT id, post_id, vote_score, edited, trending_score, content, parent_root, created_at, updated_at, hidden,
					user_id, children_count, downvote, upvote, numerical_user, numerical_replying_user,
					numerical_replying_user_is_op, numerical_user_is_op FROM top_root_comments
				UNION ALL
				SELECT id, post_id, vote_score, edited, trending_score, content, parent_root, created_at, updated_at, hidden,
					user_id, children_count, downvote, upvote, numerical_user, numerical_replying_user,
					numerical_replying_user_is_op, numerical_user_is_op
				FROM ranked_replies
				WHERE reply_num <= ?
			) AS combined_comments
		) AS t;
    `, postID, config.RootCommentsLoadedInitially, uid, config.RepliesLoadedInitially).
		Find(&comments)

	if query.Error != nil {
		return nil, query.Error
	}

	parentMap := make(map[int][]CommentDetail) // Map to store parent comments
	for i := range comments {
		comment := &comments[i]
		if !utils.ProfanityEnabled(c) {
			comment.Comment = comment.Comment.CensorComment()
		}
		comment.Comment.ObscureIfHidden()
		if comment.Comment.ParentRoot != nil {
			// aka, it's a reply
			parentID := comment.Comment.ParentRoot
			parentMap[int(*parentID)] = append(parentMap[int(*parentID)], *comment)
		} else {
			id := fmt.Sprint(comment.Comment.ID)
			err := h.redis.SAdd(c, commentSpecificKey, id).Err()
			if err != nil {
				logger.StdErr(err)
				return nil, errors.New("failed to update redis cache")
			}
		}

	}

	// Create the final list of comment threads
	var commentThreads []CommentThreadGroup
	for _, comment := range comments {
		if comment.Comment.ParentRoot == nil {
			thread := CommentThreadGroup{
				Root:    comment,
				Replies: parentMap[int(comment.Comment.ID)],
			}

			// Set the Next cursor for the last thread
			if len(thread.Replies) > 0 {
				lastReply := thread.Replies[len(thread.Replies)-1]
				time := lastReply.Comment.CreatedAt.MicroSeconds()
				thread.Next = &time
			} else {

			}
			if len(thread.Replies) == 0 {
				thread.Replies = []CommentDetail{}
			}
			commentThreads = append(commentThreads, thread)
		}
	}

	return commentThreads, nil

}

func (h *handler) handleGetComments(c *gin.Context) {
	// extract request

	var req validation.InitialCommentQuery
	err := utils.New(c).Validate(&req)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// session key that can only be created by *this* user, so it can't be guessed to manipulate others' feeds
	commentSpecificKey, err := utils.CreateCacheKey(config.RedisCommentsCache, token.UID, req.SessionKey)
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

	// retrieve the seen root comment IDs from the cache
	ids, err := h.redis.SMembers(c, commentSpecificKey).Result()
	if err != nil {
		if err == redis.Nil {
			ids = []string{} // assigns an empty slice
		} else {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	// fetch comments using the translated SQL query
	comments, err := fetchComments(int64(req.PostID), h.db, ids, req.Sort, token.UID, *h, c, commentSpecificKey)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// set the expiration for the cache
	err = h.redis.Expire(c, commentSpecificKey, seenCommentsCacheExpiry).Err()
	if err != nil {
		logger.StdErr(err)
		response.New(http.StatusInternalServerError).Err("failed to set cache expiration").Send(c)
		return
	}

	if len(comments) == 0 {
		comments = []CommentThreadGroup{}
	}

	// for each thread comment group if replies is empty make it []:
	for i := range comments {
		if len(comments[i].Replies) == 0 {
			comments[i].Replies = []CommentDetail{}
		}
	}

	// Send response
	response.New(http.StatusOK).Val(comments).Send(c)
}
