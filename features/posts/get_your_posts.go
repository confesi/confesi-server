package posts

import (
	"confesi/config"
	tags "confesi/lib/emojis"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *handler) handleGetYourPosts(c *gin.Context) {
	// extract request
	var req validation.YourPostsQuery
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	next := time.UnixMicro(int64(req.Next))

	posts := []PostDetail{}
	err = h.db.
		Preload("School").
		Preload("Faculty").
		Where("user_id = ?", token.UID).
		Where("created_at < ?", next).
		Where("hidden = ?", false).
		Order("created_at DESC").
		Find(&posts).
		Limit(config.YourPostsPageSize).
		Error

	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	for i := range posts {
		// create ref to post
		post := &posts[i]
		if post.UserID == token.UID {
			post.Owner = true
		}
		post.Emojis = tags.GetEmojis(&post.Post)
	}

	response.New(http.StatusOK).Val(posts).Send(c)
}
