package drafts

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
	"net/http"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

func (h *handler) createDraft(c *gin.Context, title string, body string, token *auth.Token) error {

	// draft to save to postgres
	draft := db.Draft{
		UserID:  token.UID,
		Title:   title,
		Content: body,
		// `HottestOn` not included so that it defaults to NULL
	}

	// save user to postgres
	err := h.db.Create(&draft).Error
	if err != nil {
		return errors.New(serverError.Error())
	}

	return nil
}

func (h *handler) handleCreate(c *gin.Context) {

	// extract request
	var req validation.CreateDraftDetails
	err := utils.New(c).ForceCustomTag("required_without", validation.RequiredWithout).Validate(&req)
	if err != nil {
		return
	}

	// strip whitespace from title and body (custom validator already confirmed this is still not empty)
	title := strings.TrimSpace(req.Title)
	body := strings.TrimSpace(req.Body)

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err("server error").Send(c)
		return
	}

	err = h.createDraft(c, title, body, token)
	if err != nil {
		response.New(http.StatusBadRequest).Err(err.Error()).Send(c)
		return
	}

	// if all goes well, send 201
	response.New(http.StatusCreated).Send(c)
}
