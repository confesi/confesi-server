package logger

import (
	"fmt"
	"time"
)

func StdInfo(s string) {
	now := time.Now()
	year, month, date := now.UTC().Date()
	hour := now.UTC().Hour()
	minute := now.UTC().Minute()
	fmt.Printf("%v-%v-%v %v:%v: %s\n", month, date, year, hour, minute, s)
}
