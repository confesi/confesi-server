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
)

// todo: check if content matching id actually exists before adding
// todo: utils/middleware for grabbing user id from token

func (h *handler) doVote(c *gin.Context, vote db.Vote) error {
	// start a transaction
	tx := h.db.Begin()
	// if something goes ary, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			// todo: will this always trigger or just for "transaction-specific" errors?
			// todo: AKA, make specific errors for each unique case?
			response.New(http.StatusInternalServerError).Err("server error").Send(c)
			return
		}
	}()
	// save the vote
	// if the vote doesn't exist, create it, else, update it

	// err := tx.Where(...)

	err := tx.FirstOrCreate(&vote).Error
	if err != nil {
		tx.Rollback()
		return errors.New("server error")
	}
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

	// update vote, if unique, then update CONTENT with new vote count and call hooks to run delta score
	// and trending algorithms

}
