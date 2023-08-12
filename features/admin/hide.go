package admin

import (
	"confesi/config/builders"
	"confesi/db"
	"confesi/lib/encryption"
	"confesi/lib/logger"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"fmt"
	"net/http"

	fcm "confesi/lib/firebase_cloud_messaging"

	"github.com/gin-gonic/gin"
)

type fcmTokenWithReportID struct {
	Token    string `gorm:"column:token"`
	ReportID uint   `gorm:"column:report_id"`
}

type fcmTokenWithOffendingHideLogID struct {
	Token     string `gorm:"column:token"`
	HideLogID uint   `gorm:"column:hide_log_id"`
}

func (h *handler) handleHideContent(c *gin.Context) {

	// validate request json
	var req validation.HideContent
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	unmaskedContentId, err := encryption.Unmask(req.ContentID)
	if err != nil {
		response.New(http.StatusBadRequest).Err(invalidValue.Error()).Send(c)
		return
	}

	hideLogEntry := db.HideLog{}
	var commentOrPostIdMatcher string

	var table string
	if req.ContentType == "comment" {
		table = "comments"
		hideLogEntry.CommentID = &db.EncryptedID{Val: unmaskedContentId}
		commentOrPostIdMatcher = "comment_id"
	} else if req.ContentType == "post" {
		table = "posts"
		hideLogEntry.PostID = &db.EncryptedID{Val: unmaskedContentId}
		commentOrPostIdMatcher = "post_id"
	} else {
		response.New(http.StatusBadRequest).Err(invalidValue.Error()).Send(c)
		return
	}

	// start a transaction
	tx := h.db.Begin()
	// if something goes ary, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}()

	updateData := map[string]interface{}{
		"hidden":          req.Hide,
		"reviewed_by_mod": req.ReviewedByMod,
	}

	// update the "hidden" field on content.
	result := tx.
		Table(table).
		Where("id = ?", req.ContentID).
		Updates(updateData)

	if result.Error != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		response.New(http.StatusBadRequest).Err(notFound.Error()).Send(c)
		return
	}

	// update all the reports for this content
	err = tx.
		Table("reports").
		Where(commentOrPostIdMatcher+" = ?", req.ContentID).
		Updates(map[string]interface{}{
			"has_been_removed": req.Hide,
			"result":           req.Reason,
		}).
		Error

	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// get the offending user's user_id
	var offendingContentUserId string
	err = tx.
		Table(table).
		Where("id = ?", req.ContentID).
		Pluck("user_id", &offendingContentUserId).
		Error

	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// create hide_log entry
	hideLogEntry.Reason = req.Reason
	hideLogEntry.UserID = offendingContentUserId
	hideLogEntry.Removed = *req.Hide

	// save it
	err = tx.Create(&hideLogEntry).
		Error

	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// commit the transaction
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// return early! just keep working on the fcm messages in the background
	response.New(http.StatusOK).Send(c)

	var reports []fcmTokenWithReportID
	var offenders []fcmTokenWithOffendingHideLogID

	// notify users who reported about the change
	err = h.db.
		Table("fcm_tokens").
		Select("fcm_tokens.token, reports.id as report_id").
		Joins("JOIN users ON users.id = fcm_tokens.user_id").
		Joins("JOIN reports ON reports.reported_by = users.id").
		Joins("JOIN "+table+" ON reports."+commentOrPostIdMatcher+" = "+table+".id").
		Where(table+".id = ?", req.ContentID).
		Scan(&reports).
		Error
	// (ignore errors, just log)
	if err != nil {
		logger.StdInfo(fmt.Sprintf("error while fetching tokens for reports: %s", err))
	} else if len(reports) > 0 {
		for _, tokenWithReportID := range reports {
			fcm.New(h.fb.MsgClient).
				ToTokens([]string{tokenWithReportID.Token}).
				WithMsg(builders.HideReportNoti()).
				WithData(builders.HideReportData(tokenWithReportID.ReportID)).
				Send(*h.db)
		}
	}

	// notify the offending user about the change
	err = h.db.
		Table("fcm_tokens").
		Select("fcm_tokens.token, hide_log.id as hide_log_id").
		Joins("JOIN users ON users.id = fcm_tokens.user_id").
		Joins("JOIN hide_log ON hide_log.user_id = users.id").
		Scan(&offenders).
		Error
	// (ignore errors, just log)
	if err != nil {
		logger.StdInfo(fmt.Sprintf("error while fetching tokens for offending user: %s", err))
	} else if len(offenders) > 0 {
		for _, tokenWithOffenderID := range offenders {
			go fcm.New(h.fb.MsgClient).
				ToTokens([]string{tokenWithOffenderID.Token}).
				WithMsg(builders.HideOffendingUserNoti()).
				WithData(builders.HideOffendingUserData(tokenWithOffenderID.HideLogID)).
				Send(*h.db)
		}
	}
}
