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

func ThreadedCommentReply(comment string) *messaging.Notification {
	truncatedComment := comment
	if len(comment) > truncationCommentLength {
		truncatedComment = comment[:truncationCommentLength]
	}
	return &messaging.Notification{
		Title: "Someone replied to your comment",
		Body:  "💬 " + truncatedComment,
	}
}

func CommentAddedToPost(comment string) *messaging.Notification {
	truncatedComment := comment
	if len(comment) > truncationCommentLength {
		truncatedComment = comment[:truncationCommentLength]
	}
	return &messaging.Notification{
		Title: "Someone commented on your confession",
		Body:  "💬 " + truncatedComment,
	}
}

func VoteOnComment(comment string, vote int) *messaging.Notification {
	truncatedComment := comment
	if len(comment) > truncationCommentLength {
		truncatedComment = comment[:truncationCommentLength]
	}
	var emoji string
	var voteString string
	if vote > 0 {
		emoji = "👍"
		voteString = "upvoted"
	} else {
		emoji = "👎"
		voteString = "downvoted"
	}
	return &messaging.Notification{
		Title: "Someone " + voteString + " your comment",
		Body:  emoji + " " + truncatedComment,
	}
}

func VoteOnPost(comment string, vote int) *messaging.Notification {
	truncatedComment := comment
	if len(comment) > truncationCommentLength {
		truncatedComment = comment[:truncationCommentLength]
	}
	var emoji string
	var voteString string
	if vote > 0 {
		emoji = "👍"
		voteString = "upvoted"
	} else {
		emoji = "👎"
		voteString = "downvoted"
	}
	return &messaging.Notification{
		Title: "Someone " + voteString + " your confession",
		Body:  emoji + " " + truncatedComment,
	}
}

func DailyHottestPosts(n int, date string) *messaging.Notification {
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

func TrendingPost(post db.Post) *messaging.Notification {

	truncatedContent := post.Title + "\n" + post.Content
	if len(post.Content) > truncatedPostTitleAndBodyLength {
		truncatedContent = post.Content[:truncatedPostTitleAndBodyLength]
	}
	return &messaging.Notification{
		Title: "A new confession is trending",
		Body:  "🚀 " + truncatedContent,
	}
}
