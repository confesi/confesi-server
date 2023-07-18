package posts

import (
	tags "confesi/lib/emojis"
	"confesi/lib/response"
	"confesi/lib/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetPostByID godoc
//
//	@Summary		Get Post By ID.
//	@Description	Fetch a post by id.
//	@Tags			Posts
//	@Accept			application/json
//	@Produce		application/json
//	@Security		BearerAuth
//	@Security		X-AppCheck-Token
//
//	@Param			date	query		string				true	"Example: 27"
//
//	@Success		200		{object}	docs.PostFound		"Post Found"
//	@Failure		400		{object}	docs.PostNotFound	"Post was Not Found"
//	@Failure		410		{object}	docs.PostRemoved	"Post was Removed"
//	@Failure		500		{object}	docs.ServerError	"Server Error"
//
//	@Router			/posts/post [get]
func (h *handler) handleGetPostById(c *gin.Context) {
	postID := c.Query("id")
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	var post PostDetail

	err = h.db.
		Preload("School").
		Preload("Faculty").
		Raw(`
				SELECT posts.*, 
					COALESCE(
						(SELECT votes.vote
						FROM votes
						WHERE votes.post_id = posts.id
							AND votes.user_id = ?
						LIMIT 1),
						'0'::vote_score_value
					) AS user_vote
				FROM posts
				WHERE posts.id = ?
				LIMIT 1
			`, token.UID, postID).
		First(&post).
		Error

	// check if the user is the owner of the post
	if post.UserID == token.UID {
		post.Owner = true
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.New(http.StatusBadRequest).Err("post not found").Send(c)
			return
		}
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	if post.Hidden {
		response.New(http.StatusGone).Err("post removed").Send(c)
		return
	}
	post.Emojis = tags.GetEmojis(&post.Post)
	response.New(http.StatusOK).Val(post).Send(c)
	return
}
