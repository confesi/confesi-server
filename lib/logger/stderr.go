package logger

import (
	"fmt"
	"os"
	"time"
)

func StdErr(m error) {
	now := time.Now()
	year, month, date := now.UTC().Date()
	hour := now.UTC().Hour()
	minute := now.UTC().Minute()

	str := fmt.Sprintf("%v-%v-%v %v:%v: ", month, date, year, hour, minute) + m.Error()
	os.Stderr.Write([]byte(str))
}
