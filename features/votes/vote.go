package votes

import (
	"confesi/db"
	"confesi/lib/algorithm"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type contentMatcher struct {
	fieldName string
	id        *uint
	model     interface{}
}

func voteToColumnName(vote int) (string, error) {
	switch vote {
	case 1:
		return "upvote", nil
	case -1:
		return "downvote", nil
	default:
		return "", invalidValue
	}
}

func (h *handler) doVote(c *gin.Context, vote db.Vote, contentType string) error {
	// start a transaction
	tx := h.db.Begin()
	// if something goes ary, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}()

	var content contentMatcher
	if contentType == "comment" {
		content = contentMatcher{fieldName: "comment_id", id: &vote.CommentID, model: &db.Comment{}}
	} else if contentType == "post" {
		content = contentMatcher{fieldName: "post_id", id: &vote.PostID, model: &db.Post{}}
	} else {
		tx.Rollback()
		return invalidValue
	}

	var oldVote int

	var model db.Vote
	// find if there's an existing vote matching id and user and content type
	if err := tx.Model(&model).Where(content.fieldName+" = ? AND user_id = ?", content.id, vote.UserID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			oldVote = 0
		} else {
			tx.Rollback()
			return serverError
		}
	} else {
		oldVote = model.Vote
	}
	// if the vote are the same, just rollback & return, there's no more work to do, but
	// we consider it idempotently a "success"
	if oldVote == vote.Vote {
		tx.Rollback()
		return nil
	}

	// if the vote value is 0, just delete it, no point storing a 0-vote
	var err error
	if vote.Vote == 0 {
		err = tx.Where(content.fieldName+" = ? AND user_id = ?", content.id, vote.UserID).Delete(&model).Error
	} else {
		// else, update/add the vote
		err = tx.
			Model(&model).
			Where(content.fieldName+" = ? AND user_id = ?", content.id, vote.UserID).
			Update("vote", vote.Vote).
			FirstOrCreate(&vote).
			Error
	}
	if err != nil {
		tx.Rollback()
		var pgErr *pgconn.PgError
		// Gorm doesn't properly handle some errors: https://github.com/go-gorm/gorm/issues/4037
		if ok := errors.As(err, &pgErr); !ok {
			// if it's not a PostgreSQL error, return a generic server error
			return serverError
		}
		switch pgErr.Code {
		case "23503": // foreign key constraint violation
			return invalidValue // aka, you provided an invalid post/comment id to try saving
		default:
			// some other postgreSQL error
			return serverError
		}
	}

	columnUpdates := make(map[string]interface{})
	var oldVoteColumn string
	var newVoteColumn string
	if oldVote != 0 {
		if oldVoteColumn, err = voteToColumnName(oldVote); err == nil {
			columnUpdates[oldVoteColumn] = gorm.Expr(oldVoteColumn+" - ?", 1)
		} else {
			tx.Rollback()
			return invalidValue
		}
	}
	if vote.Vote != 0 {
		if newVoteColumn, err = voteToColumnName(vote.Vote); err == nil {
			columnUpdates[newVoteColumn] = gorm.Expr(newVoteColumn+" + ?", 1)
		} else {
			tx.Rollback()
			return invalidValue
		}
	}

	type foundVotes struct {
		Upvote   int
		Downvote int
	}

	var votes foundVotes
	// update the score of the content
	query := tx.Model(&content.model).
		Where("id = ?", content.id).
		Updates(columnUpdates).
		Clauses(clause.Returning{}).
		Select("upvote, downvote").
		Scan(&votes)
	if err := query.Error; err != nil {
		tx.Rollback()
		return err
	}

	// update the post/comment with the modified vote values and the new trending score
	err = tx.Model(&content.model).
		Where("id = ?", content.id).
		Updates(map[string]interface{}{
			"vote_score":     votes.Upvote - votes.Downvote,                                                        // new overall content score
			"trending_score": algorithm.TrendingScore(votes.Upvote, votes.Downvote, int(time.Now().Unix()), false), // new overall trending score
		}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// commit the transaction
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return serverError
	}
	return nil
}

func (h *handler) handleVote(c *gin.Context) {
	// extract request
	var req validation.VoteDetail
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	var vote db.Vote
	vote.UserID = token.UID
	vote.Vote = int(*req.Value)
	if req.ContentType == "post" {
		vote.PostID = req.ContentID
	} else if req.ContentType == "comment" {
		vote.CommentID = req.ContentID
	} else {
		// should never happen with validated struct, but to be defensive
		response.New(http.StatusBadRequest).Err(fmt.Sprintf("invalid content type")).Send(c)
		return
	}

	if err := h.doVote(c, vote, req.ContentType); err != nil {
		// errors are always server error if they arise from here
		switch err {
		case invalidValue:
			response.New(http.StatusBadRequest).Err("invalid value").Send(c)
		default:
			response.New(http.StatusInternalServerError).Err("server error").Send(c)
		}
		return
	}

	response.New(http.StatusOK).Send(c)
}
