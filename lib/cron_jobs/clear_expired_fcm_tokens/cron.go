package clear_expired_fcm_tokens

import (
	"confesi/db"
	"confesi/lib/cron_jobs"
	"confesi/lib/logger"
	"time"

	"github.com/go-co-op/gocron"
)

func StartClearExpiredFcmTokensCronJob() {
	s := gocron.NewScheduler(time.UTC)
	s.Every(15).Day().At("00:00").Do(func() {
		cron_jobs.RetryLoop(1000, 1000*60, 24.0, 20, func() error {
			return DoClearExpiredFcmTokenJob(time.Now().UTC())
		})
	})
	logger.StdInfo("started scheduler for clear expired fcm tokens cron job")
	s.StartAsync()
}

func DoClearExpiredFcmTokenJob(dateTime time.Time) error {
	dbConn := db.New()

	twoMonthsAgo := dateTime.AddDate(0, -2, 0) // calculate the time 2 months ago

	if err := dbConn.
		Delete(&db.FcmToken{}, "updated_at < ?", twoMonthsAgo).
		Error; err != nil {
		return err
	}

	return nil
}
