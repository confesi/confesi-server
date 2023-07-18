package posts

import (
	"confesi/config"
	tags "confesi/lib/emojis"
	"confesi/lib/response"
	"confesi/lib/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetHottest godoc
//
//	@Summary		Get Hottest Posts.
//	@Description	Fetch hottest posts from a specific day
//	@Tags			Posts
//	@Accept			application/json
//	@Produce		application/json
//	@Security		BearerAuth
//	@Security		X-AppCheck-Token
//
//	@Param			date	query		string					true	"Date Example: 2023-07-04"
//
//	@Success		200		{object}	docs.HottestPosts		"Hottest"
//	@Failure		500		{object}	docs.ServerError		"Server Error"
//	@Failure		400		{object}	docs.InvalidDateFormat	"Invalid Date Format"
//
//	@Router			/posts/hottest [get]
func (h *handler) getHottestPosts(c *gin.Context, date time.Time, userID string) ([]PostDetail, error) {
	var posts []PostDetail
	err := h.db.
		Where("hottest_on = ?", date).
		Where("hidden = ?", false).
		Limit(config.HottestPostsPageSize).
		Preload("School").
		Preload("Faculty").
		Order("trending_score DESC").
		Select(`
			posts.*,
			COALESCE(
				(
					SELECT votes.vote
					FROM votes
					WHERE votes.post_id = posts.id
					AND votes.user_id = ?
					LIMIT 1
				),
				'0'::vote_score_value
			) AS user_vote
		`, userID).
		Find(&posts).
		Error
	if err != nil {
		return nil, serverError
	}
	return posts, nil
}

func (h *handler) handleGetHottest(c *gin.Context) {

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	dateStr := c.Query("day")

	// Parse the date string into a time.Time value
	date, err := time.Parse("2006-01-02", dateStr) // this basically says YYYY-MM-DD, not sure why, but it only works with a dummy date example?
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid date format").Send(c)
		return
	}

	posts, err := h.getHottestPosts(c, date, token.UID)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		return
	}

	for i := range posts {
		post := &posts[i]
		// check if the user is the owner of each post
		if post.UserID == token.UID {
			post.Owner = true
		}
		post.Emojis = tags.GetEmojis(&post.Post)
	}

	response.New(http.StatusOK).Val(posts).Send(c)
}
