package posts

import (
	"confesi/db"
	"confesi/lib/emojis"
	"confesi/lib/encryption"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleEditPost(c *gin.Context) {
	// validate the json body from request
	var req validation.EditPost
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	// get user token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	unmaskedID, err := encryption.Unmask(req.PostID)
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid id").Send(c)
		return
	}

	// sentiment analysis
	sentiment := AnalyzeText(req.Title + "\n" + req.Body)
	sentimentValue := sentiment.Compound
	if sentimentValue == 0 {
		sentimentValue = sentiment.Neutral
	}

	updates := map[string]interface{}{
		"edited":    true,
		"title":     req.Title,
		"content":   req.Body,
		"sentiment": &sentimentValue,
	}

	var post PostDetail

	// Update the `Title`/`Body` and `Edited` fields of the post in a single query
	results := h.db.
		Select(`
			posts.*, COALESCE(
				(
					SELECT votes.vote
					FROM votes
					WHERE votes.post_id = posts.id
					AND votes.user_id = ?
					LIMIT 1
				),
				'0'::vote_score_value
			) AS user_vote`, token.UID).
		Model(&db.Post{}).
		Where("id = ?", unmaskedID).
		Where("hidden = false").
		Where("user_id = ?", token.UID).
		Updates(updates).
		Preload("School").
		Preload("YearOfStudy").
		Preload("Category").
		Preload("Faculty").
		Find(&post)

	if results.Error != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	if results.RowsAffected == 0 {
		response.New(http.StatusNotFound).Err(notFound.Error()).Send(c)
		return
	}
	response.New(http.StatusOK).Val(PostDetail{Post: post.Post, Owner: true, UserVote: post.UserVote, Emojis: emojis.GetEmojis(&post.Post)}).Send(c)
}
