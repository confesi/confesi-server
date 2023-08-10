package comments

import (
	"confesi/lib/masking"
	"confesi/lib/response"
	"confesi/lib/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *handler) handleGetCommentById(c *gin.Context) {
	commentID := c.Query("id")

	unmaskedId, err := masking.Unmask(commentID)
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid id").Send(c)
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	var comment CommentDetail

	err = h.db.
		Raw(`
				SELECT comments.*,
					COALESCE(
						(SELECT votes.vote
						FROM votes
						WHERE votes.comment_id = comments.id
							AND votes.user_id = ?
						LIMIT 1),
						'0'::vote_score_value
					) AS user_vote
				FROM comments
				WHERE comments.id = ?
				LIMIT 1
			`, token.UID, unmaskedId).
		First(&comment).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.New(http.StatusBadRequest).Err("comment not found").Send(c)
			return
		}
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	if comment.Comment.Hidden {
		response.New(http.StatusGone).Err("comment removed").Send(c)
		return
	}
	// check if user is owner
	if comment.Comment.UserID == token.UID {
		comment.Owner = true
	}
	if !utils.ProfanityEnabled(c) {
		comment.Comment = comment.Comment.CensorComment()
	}
	response.New(http.StatusOK).Val(comment).Send(c)
}
