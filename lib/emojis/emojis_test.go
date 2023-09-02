package emojis

import (
	"confesi/db"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

//! Tests require `MASK_SECRET` env var to be set to pass

func TestHottestEmoji(t *testing.T) {
	emojis := GetEmojis(&db.Post{HottestOn: &datatypes.Date{}})
	assert.Equal(t, []string{"🔥"}, emojis, "emojis do not match")
}

func TestMaxLengthEmoji(t *testing.T) {
	emojis := GetEmojis(&db.Post{Content: string(make([]byte, 2000))})
	assert.Equal(t, []string{"💬"}, emojis, "emojis do not match")
}

func TestDislikedEmoji(t *testing.T) {
	emojis := GetEmojis(&db.Post{Downvote: 1, Upvote: 0})
	assert.Equal(t, []string{"💀"}, emojis, "emojis do not match")
}

func TestWellLikedEmoji(t *testing.T) {
	emojis := GetEmojis(&db.Post{Upvote: 100})
	assert.Equal(t, []string{"💙"}, emojis, "emojis do not match")
}

func TestWellDislikedEmoji(t *testing.T) {
	emojis := GetEmojis(&db.Post{Downvote: 100, Upvote: 0})
	assert.Equal(t, []string{"💀", "🤡"}, emojis, "emojis do not match")
}

func TestMultipleEmoji(t *testing.T) {
	emojis := GetEmojis(&db.Post{Downvote: 300, Upvote: 100, Content: string(make([]byte, 2000))})
	assert.Equal(t, []string{"💬", "💀", "💙", "🤡"}, emojis, "emojis do not match")
}
