package builders

import (
	"confesi/db"
	"fmt"
	rnd "math/rand"

	"firebase.google.com/go/messaging"
)

const (
	truncationCommentLength         = 100
	truncatedPostTitleAndBodyLength = 150
)

func ThreadedCommentReplyNoti(comment string) *messaging.Notification {
	truncatedComment := comment
	if len(comment) > truncationCommentLength {
		truncatedComment = comment[:truncationCommentLength]
	}
	return &messaging.Notification{
		Title: "Someone replied to your comment",
		Body:  "💬 " + truncatedComment,
	}
}

func CommentAddedToPostNoti(comment string) *messaging.Notification {
	truncatedComment := comment
	if len(comment) > truncationCommentLength {
		truncatedComment = comment[:truncationCommentLength]
	}
	return &messaging.Notification{
		Title: "Someone commented on your confession",
		Body:  "💬 " + truncatedComment,
	}
}

func VoteOnCommentNoti(vote int, totalVotes int) *messaging.Notification {
	var voteString string
	var emoji string
	if vote > 0 {
		emoji = "👍"
		voteString = "upvote"
	} else {
		emoji = "👎"
		voteString = "downvote"
	}
	return &messaging.Notification{
		Title: "New " + voteString + " on your comment " + emoji,
		Body:  fmt.Sprintf("Total: %d", totalVotes),
	}
}

func VoteOnPostNoti(vote int, totalVotes int) *messaging.Notification {
	var voteString string
	var emoji string
	if vote > 0 {
		emoji = "👍"
		voteString = "upvote"
	} else {
		emoji = "👎"
		voteString = "downvote"
	}
	return &messaging.Notification{
		Title: "New " + voteString + " " + emoji,
		Body:  fmt.Sprintf("Total: %d", totalVotes),
	}
}

func YourSchoolsDailyHottestNoti(occurences int) *messaging.Notification {
	if occurences == 1 {
		return &messaging.Notification{
			Title: "Daily Hottest",
			Body:  fmt.Sprintf("🔥 Your school reached the Daily Hottest page today"),
		}
	}

	return &messaging.Notification{
		Title: "Daily Hottest",
		Body:  fmt.Sprintf("🔥 Your school reached the Daily Hottest page %d times today", occurences),
	}
}

func DailyHottestPostsNoti(n int, date string) *messaging.Notification {
	var confessionWord string
	if n == 1 {
		confessionWord = "confession"
	} else {
		confessionWord = "confessions"
	}
	return &messaging.Notification{
		Title: fmt.Sprintf("Check out today's hottest %d %s", n, confessionWord),
		Body:  "🔥 " + date,
	}
}

func TrendingPostNoti(post db.Post) *messaging.Notification {

	truncatedContent := post.Title + "\n" + post.Content
	if len(post.Content) > truncatedPostTitleAndBodyLength {
		truncatedContent = post.Content[:truncatedPostTitleAndBodyLength]
	}
	return &messaging.Notification{
		Title: "A new confession is trending",
		Body:  "🚀 " + truncatedContent,
	}
}

func HideOffendingUserNoti() *messaging.Notification {
	return &messaging.Notification{
		Title: "⛔️ Moderator action",
		Body:  "A comment or confession you created was deleted by a mod. You can view the reason, if given, in-app.",
	}
}

func HideReportNoti() *messaging.Notification {
	return &messaging.Notification{
		Title: "✅ Moderator action",
		Body:  "Something you reported was deleted by a mod. Thanks for your input. You can view the report results, if given, in-app.",
	}
}

func AdminSendNotificationNoti(title string, body string) *messaging.Notification {
	return &messaging.Notification{
		Title: title,
		Body:  body,
	}
}

func AccountStandingBannedNoti(isBanned bool) *messaging.Notification {
	var word string
	if isBanned {
		word = "banned"
	} else {
		word = "unbanned"
	}
	return &messaging.Notification{
		Title: "🎱 Your account has been " + word,
		Body:  "Contact support for more information",
	}
}

func AccountStandingLimitedNoti(isLimited bool) *messaging.Notification {
	var word string
	if isLimited {
		word = "limited"
	} else {
		word = "fully re-enabled"
	}
	return &messaging.Notification{
		Title: "🎱 Your account has been " + word,
		Body:  "Contact support for more information",
	}
}

func NewRoomCreatedNoti() *messaging.Notification {
	return &messaging.Notification{
		Title: "💬 Somebody created a chat room with you",
		Body:  "Check it out",
	}
}

func NewChatNoti(message string, room string) *messaging.Notification {
	return &messaging.Notification{
		Title: room,
		Body:  message,
	}
}

func DeletedChatNoti() *messaging.Notification {
	return &messaging.Notification{
		Title: genRandSusEmoji() + " Somebody deleted a chat",
		Body:  "A typo? A regret? Something sus? Who knows...",
	}
}

func genRandSusEmoji() string {
	possibleEmojis := []string{
		"👀",
		"👁",
		"😏",
		"🤔",
		"🤨",
		"🫣",
		"🥶",
	}
	// rand idx using math/random lib
	randIdx := rnd.Intn(len(possibleEmojis))
	return possibleEmojis[randIdx]
}
