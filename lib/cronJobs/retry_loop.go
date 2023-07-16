package cronJobs

import (
	"confesi/lib/logger"
	"errors"
	"strconv"
	"time"
)

// [initialDelayMs]: the initial delay duration in milliseconds
// [additionalDelayPerAttemptMs]: delay factor multiplier value
// [hoursAfterOriginalAttemptToRetryFor] and [maxRetries]: whichever limit hits first, we stop
func RetryLoop(initialDelayMs int, additionalDelayPerAttemptMs int, hoursAfterOriginalAttemptToRetryFor float64, maxRetries int, executeCronJob func() error) {
	logger.StdInfo("starting daily hottest posts cron job")
	dateTime := time.Now().UTC()
	baseDelayMs := initialDelayMs
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := executeCronJob()
		if err != nil {
			// job failed, so retry
			logger.StdErr(errors.New("daily hottest posts cron job errored on attempt (attempt " + strconv.Itoa(attempt) + ", after " + time.Since(dateTime).String() + ")"))
			delay := time.Duration(baseDelayMs) * time.Millisecond
			if time.Since(dateTime)+delay > time.Duration(hoursAfterOriginalAttemptToRetryFor)*time.Hour {
				time.Sleep(time.Duration(hoursAfterOriginalAttemptToRetryFor)*time.Hour - time.Since(dateTime))
			} else {
				time.Sleep(delay)
			}
			baseDelayMs += additionalDelayPerAttemptMs
			// if we're past a certain preset number of hours, or are passed our maxRetries, then give up
			if time.Since(dateTime).Hours() > hoursAfterOriginalAttemptToRetryFor || attempt > maxRetries {
				// job failed
				logger.StdErr(errors.New("daily hottest posts cron job failed and exited (attempt " + strconv.Itoa(attempt) + ", after " + time.Since(dateTime).String() + ")"))
				break
			}
		} else {
			// job done successfully!
			logger.StdInfo("daily hottest posts cron job done successfully")
			break
		}
	}
}
