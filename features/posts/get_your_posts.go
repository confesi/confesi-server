package posts

import (
	"confesi/config"
	tags "confesi/lib/emojis"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

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

	posts := []PostDetail{}
	err = h.db.
		Preload("School").
		Preload("Faculty").
		Where("user_id = ?", token.UID).
		Raw(req.Next.Cursor("created_at >")).
		Where("hidden = ?", false).
		Order("created_at ASC").
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
