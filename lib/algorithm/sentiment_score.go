package algorithm

import "math"

// Desmos Implementation https://www.desmos.com/calculator/tsqi2sdgef

// Calculates the sentiment score of a bit of content.
//
// Accepts time in seconds since unix epoch, UTC.
func SentimentScore(sentimentScore float64, upvotes int, downvotes int, currTime int, limited bool) float64 {
	// calculate the time difference
	timeDelta := float64(currTime - baseTime)

	// calculate the total popularity
	popularity := float64(upvotes - downvotes)

	// check if the post needs to have limited reach
	var limitFactor float64
	if limited {
		limitFactor = limitedMultiplier
	} else {
		limitFactor = 1
	}

	// calculate the z-factor
	var z float64
	if popularity >= 1 {
		z = popularity
	} else {
		z = 1
	}

	// calculate the y-factor
	y := math.Abs(float64(sentimentScore)) + 1

	// Get negativity factor
	var x float64
	if sentimentScore > 0 {
		x = 1
	} else {
		x = -1
	}

	// calculate the sentiment score with logarithmic depression
	result := (math.Log10(z)/math.Log10(20) + math.Log10(timeDelta/decayTimeDivisor) + x*math.Log2(y)*2)

	if result < 0 {
		return result
	} else {
		return result * limitFactor
	}

}
