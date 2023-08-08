package admin

import (
	"confesi/config/builders"
	"confesi/db"
	fcm "confesi/lib/firebase_cloud_messaging"
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
	userRoles, err := getUserRoles(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(err.Error()).Send(c)
		return
	}
	if len(userRoles.SchoolMods) > 0 {
		user := db.User{}
		query := h.db.Model(&db.User{}).Where("id = ?", req.UserID)

		res := query.First(&user)
		if res.Error != nil {
			response.New(http.StatusForbidden).Err("user doesn't exist").Send(c)
			return
		}

		res = query.Where("school_id IN ?", userRoles.SchoolMods).First(&user)

		if res.Error != nil {
			response.New(http.StatusForbidden).Err("missing school permissions").Send(c)
			return
		}
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

	// fcm notifications to affected users
	var tokens []string
	err = h.db.
		Table("fcm_tokens").
		Select("fcm_tokens.token").
		Joins("JOIN users ON users.id = fcm_tokens.user_id").
		Where("users.id = ?", req.UserID).
		Pluck("fcm_tokens.token", &tokens).
		Error
	if err == nil && len(tokens) > 0 {
		// don't handle errors here, because it's not a big deal if the notification doesn't send
		if req.Standing == "limited" || req.Standing == "enabled" {
			isLimited := req.Standing == "limited"
			fcm.New(h.fb.MsgClient).
				ToTokens(tokens).
				WithMsg(builders.AccountStandingLimitedNoti(isLimited)).
				WithData(builders.AccountStandingLimitedData(isLimited)).
				Send(*h.db)
		} else if req.Standing == "banned" || req.Standing == "unbanned" {
			isBanned := req.Standing == "banned"
			fcm.New(h.fb.MsgClient).
				ToTokens(tokens).
				WithMsg(builders.AccountStandingBannedNoti(isBanned)).
				WithData(builders.AccountStandingBannedData(isBanned)).
				Send(*h.db)
		}
	}

	// if all goes well (ignoring fcm, because we hope it works, but it's not critical it does), send 200
	response.New(http.StatusOK).Send(c)
}

// todo: send fcm
