package clearExpiredFcmTokens

import (
	"confesi/db"
	"confesi/lib/cronJobs"
	"confesi/lib/logger"
	"time"

	"github.com/go-co-op/gocron"
	"gorm.io/datatypes"
)

func StartClearExpiredFcmTokensCronJob() {
	s := gocron.NewScheduler(time.UTC)
	s.Every(15).Day().At("00:00").Do(func() {
		cronJobs.RetryLoop(1000, 1000*60, 24.0, 20, func() error {
			return DoClearExpiredFcmTokenJob(time.Now().UTC())
		})
	})
	logger.StdInfo("started scheduler for clear expired fcm tokens cron job")
	s.StartAsync()
}

func DoClearExpiredFcmTokenJob(dateTime time.Time) error {
	dbConn := db.New()

	// start a transaction
	tx := dbConn.Begin()

	twoMonthsAgo := dateTime.AddDate(0, -2, 0) // calculate the time 2 months ago

	if err := tx.
		Delete(&db.FcmToken{}, "updated_at < ?", twoMonthsAgo).
		Error; err != nil {
		tx.Rollback()
		return err
	}

	// set job on today ran
	err := tx.Create(&db.CronJob{Ran: datatypes.Date(dateTime), Type: cronJobs.DailyHottestCronJobLog}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// successfully commit transaction
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
