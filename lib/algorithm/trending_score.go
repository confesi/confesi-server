package algorithm

import "math"

// Based on Reddit's trending algorithm found in this article: https://moz.com/blog/reddit-stumbleupon-delicious-and-hacker-news-algorithms-exposed
// at this location: https://github.com/reddit-archive/reddit/blob/bd922104b971a5c6794b199f364a06fdf61359a2/r2/r2/lib/db/_sorts.pyx#L47

const (
	baseTime          = 1257894000 // base time in seconds (aka, the epoch for our trending calculations)
	limitedMultiplier = 0.5        // multiplier for limited posts
)

// Calculates the trending score of a bit of content.
//
// Accepts time in seconds since unix epoch, UTC.
func TrendingScore(upvotes int, downvotes int, currTime int, limited bool) float64 {
	// calculate the time difference
	timeDelta := float64(currTime - baseTime)

	// calculate the total score
	score := float64(upvotes - downvotes)

	// calculate the y-factor
	var y float64
	if score > 0 {
		y = 1
	} else if score < 0 {
		y = -1
	} else {
		y = 0
	}

	// calculate the z-factor
	var z float64
	if math.Abs(score) >= 1 {
		z = math.Abs(score)
	} else {
		z = 1
	}

	// check if the post needs to have limited reach
	var limitFactor float64
	if limited {
		limitFactor = limitedMultiplier
	} else {
		limitFactor = 1
	}

	// calculate the trending score
	result := (y*math.Log10(z) + (timeDelta / 45000))
	if result < 0 {
		return result
	} else {
		return result * limitFactor
	}
}
