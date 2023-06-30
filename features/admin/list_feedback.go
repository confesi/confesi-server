package admin

import (
	"confesi/db"
	"confesi/lib/response"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type fetchedFeedback struct {
	Feedback []db.Feedback `json:"feedback"`
	Next     *int64        `json:"next"`
	Previous *int64        `json:"previous"`
}

func (h *handler) handleListFeedback(c *gin.Context) {

	// get query params
	nextStr := c.Query("next")
	pageSizeStr := c.Query("limit")

	nextInt, err := strconv.Atoi(nextStr)
	// Error Check
	if err != nil || nextInt < 0 {
		response.New(http.StatusBadRequest).Err("invalid cursor").Send(c)
		return
	}
	next := time.UnixMilli(int64(nextInt))

	cursorSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid limit").Send(c)
		return
	}

	fetchResult := fetchedFeedback{}
	err = h.db.Model(&db.Feedback{}).Where("feedbacks.created_at > ?", next).Limit(cursorSize).Find(&fetchResult.Feedback).Order("id").Error

	if len(fetchResult.Feedback) == 0 {
		response.New(http.StatusNotFound).Err("no feedback found").Send(c)
		return
	}

	if err != nil {
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		return
	}

	timeMillis := fetchResult.Feedback[len(fetchResult.Feedback)-1].CreatedAt.UnixMilli()
	fetchResult.Next = &timeMillis
	previousTimeMillis := next.UnixMilli()
	fetchResult.Previous = &previousTimeMillis

	// if all goes well, send 200
	response.New(http.StatusOK).Val(fetchResult).Send(c)
}
