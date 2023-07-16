package builders

import (
	"confesi/db"
	"fmt"

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
		voteString = "upvoted"
	} else {
		emoji = "👎"
		voteString = "downvoted"
	}
	return &messaging.Notification{
		Title: emoji + " " + "Someone " + voteString + " your comment",
		Body:  fmt.Sprintf("Total votes: %d", totalVotes),
	}
}

func VoteOnPostNoti(vote int, totalVotes int) *messaging.Notification {
	var voteString string
	var emoji string
	if vote > 0 {
		emoji = "👍"
		voteString = "upvoted"
	} else {
		emoji = "👎"
		voteString = "downvoted"
	}
	return &messaging.Notification{
		Title: emoji + " " + "Someone " + voteString + " your confession",
		Body:  fmt.Sprintf("Total votes: %d", totalVotes),
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
