package utils

import "time"

// Convert time.Time to unix milliseconds
func UnixMs(time time.Time) int64 {
	return time.UnixMilli()
}
