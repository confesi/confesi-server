package votes

import (
	"confesi/config/builders"
	"confesi/db"
	"confesi/lib/algorithm"
	"confesi/lib/encryption"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
	"fmt"
	"net/http"
	"time"

	fcm "confesi/lib/firebase_cloud_messaging"

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

func (h *handler) doVote(c *gin.Context, vote db.Vote, contentType string, uid string) error {
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
		content = contentMatcher{fieldName: "comment_id", id: &vote.CommentID.Val, model: &db.Comment{}}
	} else if contentType == "post" {
		content = contentMatcher{fieldName: "post_id", id: &vote.PostID.Val, model: &db.Post{}}
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
		Clauses(clause.Locking{Strength: "UPDATE"}).
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

	// send fcm notifications
	if vote.Vote == 0 {
		return nil
	}

	// Retrieve tokens for either comment or post
	var tokens []string
	if vote.CommentID != nil {
		err = h.db.
			Table("fcm_tokens").
			Select("fcm_tokens.token").
			Joins("JOIN users ON users.id = fcm_tokens.user_id").
			Joins("JOIN comments ON comments.user_id = users.id").
			Where("comments.id = ? AND users.id <> ?", vote.CommentID, uid).
			Pluck("fcm_tokens.token", &tokens).
			Error
		if err == nil && len(tokens) > 0 && ((votes.Upvote+votes.Downvote)%5 == 0) {
			go fcm.New(h.fb.MsgClient).
				ToTokens(tokens).
				WithMsg(builders.VoteOnCommentNoti(vote.Vote, votes.Upvote-votes.Downvote)).
				WithData(builders.VoteOnCommentData(vote.CommentID.Val)).
				Send(*h.db)
		}
	} else if vote.PostID != nil {
		err = h.db.
			Table("fcm_tokens").
			Select("fcm_tokens.token").
			Joins("JOIN users ON users.id = fcm_tokens.user_id").
			Joins("JOIN posts ON posts.user_id = users.id").
			Where("posts.id = ? AND users.id <> ?", vote.PostID, uid).
			Pluck("fcm_tokens.token", &tokens).
			Error
		// print((votes.Upvote + votes.Downvote%5))
		if err == nil && len(tokens) > 0 && ((votes.Upvote+votes.Downvote)%5 == 0) {
			go fcm.New(h.fb.MsgClient).
				ToTokens(tokens).
				WithMsg(builders.VoteOnPostNoti(vote.Vote, votes.Upvote-votes.Downvote)).
				WithData(builders.VoteOnCommentData(vote.PostID.Val)).
				Send(*h.db)
		}
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

	unmaskedId, err := encryption.Unmask(req.ContentID)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	var vote db.Vote
	vote.UserID = token.UID
	vote.Vote = int(*req.Value)
	if req.ContentType == "post" {
		vote.PostID = &db.MaskedID{Val: unmaskedId}
	} else if req.ContentType == "comment" {
		vote.CommentID = &db.MaskedID{Val: unmaskedId}
	} else {
		// should never happen with validated struct, but to be defensive
		response.New(http.StatusBadRequest).Err(fmt.Sprintf("invalid content type")).Send(c)
		return
	}

	if err := h.doVote(c, vote, req.ContentType, token.UID); err != nil {
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
