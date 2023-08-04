package emojis

import "confesi/db"

// Applies emojis to a post
func GetEmojis(post *db.Post) []string {
	emojis := []string{}
	// hottest post
	if post.HottestOn != nil {
		emojis = append(emojis, "🔥")
	}
	// max-length post
	if len(post.Content) >= 2000 {
		emojis = append(emojis, "💬")
	}
	// disliked post
	if post.Downvote > post.Upvote {
		emojis = append(emojis, "💀")
	}
	// well-liked post
	if post.Upvote >= 100 {
		emojis = append(emojis, "💙")
	}
	// well-disliked post (int cast so not unsigned int rollover)
	if int(post.Downvote-post.Upvote) >= 100 {
		emojis = append(emojis, "🤡")
	}
	return emojis
}
