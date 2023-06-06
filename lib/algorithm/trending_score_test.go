package algorithm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests posts with different upvotes and downvote counts.
func TestTrendingDiffScore(t *testing.T) {
	// post 1
	p1Upvotes := 10             // upvotes
	p1Downvotes := 10           // downvotes
	p1CurrTime := 1686087193051 // curr time in ms
	p1Limited := false          // is limited

	// post 2
	p2Upvotes := 15             // upvotes
	p2Downvotes := 10           // downvotes
	p2CurrTime := 1686087193051 // curr time in ms
	p2Limited := false          // is limited

	p1TrendingScore := TrendingScore(p1Upvotes, p1Downvotes, p1CurrTime, p1Limited)
	p2TrendingScore := TrendingScore(p2Upvotes, p2Downvotes, p2CurrTime, p2Limited)
	assert.Less(t, p1TrendingScore, p2TrendingScore)
}

// Tests posts with the same upvotes and downvotes, but where one is "limited".
func TestTrendingLimited(t *testing.T) {
	// post 1
	p1Upvotes := 10             // upvotes
	p1Downvotes := 10           // downvotes
	p1CurrTime := 1686087193051 // curr time in ms
	p1Limited := false          // is limited

	// post 2
	p2Upvotes := 10             // upvotes
	p2Downvotes := 10           // downvotes
	p2CurrTime := 1686087193051 // curr time in ms
	p2Limited := true           // is limited

	p1TrendingScore := TrendingScore(p1Upvotes, p1Downvotes, p1CurrTime, p1Limited)
	p2TrendingScore := TrendingScore(p2Upvotes, p2Downvotes, p2CurrTime, p2Limited)
	assert.Less(t, p2TrendingScore, p1TrendingScore)
}

// Test posts created at different times, where the older one's score is higher, but
// should have decayed to a lower score than the newer post.
func TestTrendingTimeDecay(t *testing.T) {
	// post 1
	p1Upvotes := 100            // upvotes
	p1Downvotes := 10           // downvotes
	p1CurrTime := 1686000193051 // curr time in ms
	p1Limited := false          // is limited

	// post 2
	p2Upvotes := 50             // upvotes
	p2Downvotes := 10           // downvotes
	p2CurrTime := 1686087193051 // curr time in ms
	p2Limited := false          // is limited

	p1TrendingScore := TrendingScore(p1Upvotes, p1Downvotes, p1CurrTime, p1Limited)
	p2TrendingScore := TrendingScore(p2Upvotes, p2Downvotes, p2CurrTime, p2Limited)
	assert.Less(t, p1TrendingScore, p2TrendingScore)
}
