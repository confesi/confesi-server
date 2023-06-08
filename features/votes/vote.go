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

	fmt.Println("here 1")

	var content contentMatcher
	if contentType == "comment" {
		content = contentMatcher{fieldName: "comment_id", id: &vote.CommentID, model: &db.Comment{}}
	} else if contentType == "post" {
		content = contentMatcher{fieldName: "post_id", id: &vote.PostID, model: &db.Post{}}
	} else {
		tx.Rollback()
		return errors.New(InvalidContentType)
	}

	var old_vote int

	fmt.Println("here 2")

	var model db.Vote
	// find if there's an existing vote matching id and user and content type
	if err := tx.Model(&model).Where(content.fieldName+" = ? AND user_id = ?", content.id, vote.UserID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			old_vote = 0
		} else {
			tx.Rollback()
			return errors.New(ServerError)
		}
	}
	old_vote = model.Vote
	delta_vote := vote.Vote - old_vote
	fmt.Println("here 3")

	// update/create the vote
	if err := tx.Model(&model).Where(content.fieldName+" = ? AND user_id = ?", content.id, vote.UserID).FirstOrCreate(&vote).Update("vote", vote.Vote).Error; err != nil {
		tx.Rollback()
		return errors.New(ServerError)
	}

	fmt.Println("here 4")

	// todo: update each one individually, aka, upvotes, downvotes, and then hook for everything else?
	// update the score of the content
	if err := tx.Model(content.model).Where("id = ?", content.id).Update("score", gorm.Expr("score + ?", delta_vote)).Error; err != nil {
		tx.Rollback()
		return errors.New(ServerError)
	}

	fmt.Println("here 5")

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
	vote.Vote = int(req.Value)
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
