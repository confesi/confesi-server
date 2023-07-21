package admin

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

func (h *handler) handleSetUserStanding(c *gin.Context) {

	// validate request
	var req validation.UserStanding
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	if req.Standing == "limited" {
		// update a column `limited` in the `users` table to be true
		res := h.db.Model(&db.User{}).Where("id = ?", req.UserID).Update("is_limited", true)
		if res.Error != nil {
			response.New(http.StatusInternalServerError).Err("server error").Send(c)
			return
		}
		if res.RowsAffected == 0 {
			response.New(http.StatusBadRequest).Err("user doesn't exist").Send(c)
			return
		}
	} else if req.Standing == "enabled" {
		res := h.db.Model(&db.User{}).Where("id = ?", req.UserID).Update("is_limited", false)
		if res.Error != nil {
			response.New(http.StatusInternalServerError).Err("server error").Send(c)
			return
		}
		if res.RowsAffected == 0 {
			response.New(http.StatusBadRequest).Err("user doesn't exist").Send(c)
			return
		}
	} else if req.Standing == "banned" {
		update := (&auth.UserToUpdate{}).Disabled(true)
		_, err := h.fb.AuthClient.UpdateUser(c, req.UserID, update)
		if err != nil {
			response.New(http.StatusInternalServerError).Err("server error").Send(c)
			return
		}
	} else if req.Standing == "unbanned" {
		update := (&auth.UserToUpdate{}).Disabled(false)
		_, err := h.fb.AuthClient.UpdateUser(c, req.UserID, update)
		if err != nil {
			response.New(http.StatusInternalServerError).Err("server error").Send(c)
			return
		}
	} else {
		response.New(http.StatusBadRequest).Err("invalid standing").Send(c)
		return
	}

	// if all goes well, send 200
	response.New(http.StatusOK).Send(c)
}

// todo: send fcm
