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

type FetchResults struct {
	Posts []PostDetail `json:"posts"`
	Next  *int64       `json:"next"`
}

// GetYourPosts godoc
//
//	@Summary		Get Your Posts.
//	@Description	Fetch your posts.
//	@Tags			Posts
//	@Accept			application/json
//	@Produce		application/json
//	@Security		BearerAuth
//	@Security		X-AppCheck-Token
//	@Param			Body	body		string				true	"The Pagination Cursor"	SchemaExample({\n "next":1688460277629001\n})
//	@Success		201		{object}	docs.YourPosts		"Your Posts"
//	@Failure		500		{object}	docs.ServerError	"Server Error"
//
//	@Router			/posts/your-posts [get]
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

	fetchResults := FetchResults{}

	err = h.db.
		Preload("School").
		Preload("Faculty").
		Where("user_id = ?", token.UID).
		Where(req.Next.Cursor("created_at >")).
		Where("hidden = ?", false).
		Order("created_at ASC").
		Find(&fetchResults.Posts).
		Limit(config.YourPostsPageSize).
		Error

	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	if len(fetchResults.Posts) > 0 {
		timeMicros := (fetchResults.Posts[len(fetchResults.Posts)-1].CreatedAt.Time).UnixMicro()
		fetchResults.Next = &timeMicros
		for i := range fetchResults.Posts {
			// create ref to post
			post := &fetchResults.Posts[i]
			if post.UserID == token.UID {
				post.Owner = true
			}
			post.Emojis = tags.GetEmojis(&post.Post)
		}

	}

	response.New(http.StatusOK).Val(fetchResults).Send(c)
}
