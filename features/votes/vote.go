package votes

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// todo: which aren't using these? which are?
const (
	ServerError        = "server error"
	InvalidContentType = "invalid content type"
)

type contentMatcher struct {
	fieldName string
	id        *uint
	model     interface{}
}

// todo: check if content matching id actually exists before adding (FK enforces this?)
// todo: make separate vote collections for posts and comments & update schema because of double foreign keys which don't work

func (h *handler) doVote(c *gin.Context, vote db.Vote, contentType string) error {
	// start a transaction
	tx := h.db.Begin()
	// if something goes ary, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(ServerError).Send(c)
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
		return errors.New(InvalidContentType)
	}

	var oldVote int

	var model db.Vote
	// find if there's an existing vote matching id and user and content type
	if err := tx.Model(&model).Where(content.fieldName+" = ? AND user_id = ?", content.id, vote.UserID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			oldVote = 0
		} else {
			tx.Rollback()
			return errors.New(ServerError)
		}
	} else {
		oldVote = model.Vote
	}
	// if the vote are the same, just rollback & return, there's no more work to do, but
	// we consider it idempotently a "success"
	if oldVote == vote.Vote {
		fmt.Println("OLD == NEW => ROLLBACK")
		tx.Rollback()
		return nil
	}

	// if the vote value is 0, just delete it, no point storing a 0-vote
	var err error
	if vote.Vote == 0 {
		err = tx.Where(content.fieldName+" = ? AND user_id = ?", content.id, vote.UserID).Delete(&model).Error
	} else {
		// else, update/add the vote
		err = tx.Model(&model).Where(content.fieldName+" = ? AND user_id = ?", content.id, vote.UserID).Update("vote", vote.Vote).FirstOrCreate(&vote).Error
	}
	if err != nil {
		tx.Rollback()
		return errors.New(ServerError)
	}

	// todo: update each one individually, aka, upvotes, downvotes, and then hook for everything else?
	// update the score of the content
	query := tx.Model(content.model).
		Where("id = ?", content.id).
		UpdateColumns(map[string]interface{}{
			"upvote":   gorm.Expr("+1", vote.Vote, oldVote),
			"downvote": gorm.Expr("+1", vote.Vote, oldVote),
		})

	if err := query.Error; err != nil {
		tx.Rollback()
		return err
	}

	// commit the transaction
	tx.Commit()
	return nil
}

func (h *handler) handleVote(c *gin.Context) {
	// extract request
	var req validation.VoteDetail

	// validator
	binding := &validation.DefaultBinding{
		Validator: validator.New(),
	}
	if err := binding.Bind(c.Request, &req); err != nil {
		response.New(http.StatusBadRequest).Err(fmt.Sprintf("failed validation: %v", err)).Send(c)
		return
	}

	token, err := utils.UserFromContext(c)
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
		// todo: handle different types of thrown errors
		response.New(http.StatusInternalServerError).Err(fmt.Sprintf("failed to vote: %v", err)).Send(c)
		return
	}

	response.New(http.StatusOK).Send(c)
}
