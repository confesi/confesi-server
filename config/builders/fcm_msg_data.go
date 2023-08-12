package builders

import (
	"fmt"
)

func ThreadedCommentReplyData(parentCommentId uint, commentId uint, postId uint) map[string]string {
	return map[string]string{
		"type":              "threaded_comment_reply",
		"parent_comment_id": fmt.Sprintf("%v", parentCommentId),
		"comment_id":        fmt.Sprintf("%v", commentId),
		"post_id":           fmt.Sprintf("%v", postId),
	}
}

func CommentAddedToPostData(commentId uint, post_id uint) map[string]string {
	return map[string]string{
		"type":       "comment_added_to_post",
		"comment_id": fmt.Sprintf("%v", commentId),
		"post_id":    fmt.Sprintf("%v", post_id),
	}
}

func VoteOnCommentData(commentId uint) map[string]string {
	return map[string]string{
		"type":       "vote_on_comment",
		"comment_id": fmt.Sprintf("%v", commentId),
	}
}

func VoteOnPostData(postId uint) map[string]string {
	return map[string]string{
		"type":    "vote_on_post",
		"post_id": fmt.Sprintf("%v", postId),
	}
}

func DailyHottestPostsData() map[string]string {
	return map[string]string{
		"type": "daily_hottest_posts",
	}
}

func YourSchoolsDailyHottestData() map[string]string {
	return map[string]string{
		"type": "daily_hottest_posts",
	}
}

func TrendingPostData(postId uint) map[string]string {
	return map[string]string{
		"type":    "trending_post",
		"post_id": fmt.Sprintf("%v", postId),
	}
}

// todo: add proper APNS, notification data stuff here so it doesn't error out as a silent notification
func NotificationSettingsSyncData() map[string]string {
	return map[string]string{
		"type": "notification_settings_sync",
	}
}

func HideOffendingUserData(modLogId uint) map[string]string {
	return map[string]string{
		"type":       "hide_offending_user_content",
		"mod_log_id": fmt.Sprintf("%v", modLogId),
	}
}

func HideReportData(reportId uint) map[string]string {
	return map[string]string{
		"type":    "hide_report_result",
		"post_id": fmt.Sprintf("%v", reportId),
	}
}

func AccountStandingBannedData(isBanned bool) map[string]string {
	return map[string]string{
		"type":   "account_standing_banned",
		"banned": fmt.Sprintf("%v", isBanned),
	}
}

func AccountStandingLimitedData(isLimited bool) map[string]string {
	return map[string]string{
		"type":    "account_standing_limited",
		"limited": fmt.Sprintf("%v", isLimited),
	}
}
