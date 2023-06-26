package cron

import (
	"confesi/config"
	"confesi/db"
	"confesi/lib/logger"
	"errors"
	"strconv"
	"time"

	"github.com/go-co-op/gocron"
	"gorm.io/datatypes"

	"gorm.io/gorm"
)

const (
	// when to run the cron job daily in UTC
	runJobDailyAtTime = "23:55" // 11:55 PM

	// delay factor multiplier value
	additionalDelayPerAttemptMs = 1000 * 60 // 1 minute
	// initial delay duration
	initialDelayMs = 1000 // 1 second

	// whichever limit hits first, we stop
	hoursAfterOriginalAttemptToRetryFor = 6.0
	maxRetries                          = 30
)

// Cron job that runs daily to update the hottest posts.
//
// On failure, it will retry for either a certain amount of hours, or a certain number of times, whichever limit hits first.
// The delay between retries increases by a certain amount each time.
//
// If it fails after all those attempts, it will print to STDERR and give up.
func StartDailyHottestPostsCronJob() {
	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Day().At(runJobDailyAtTime).Do(cronRetryLoop)
	logger.StdInfo("started scheduler for daily hottest posts cron job")
	s.StartAsync()
}

func cronRetryLoop() {
	logger.StdInfo("starting daily hottest posts cron job")
	dateTime := time.Now().UTC()
	baseDelayMs := initialDelayMs
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := ExecuteCronJob(dateTime)
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

func ExecuteCronJob(dateTime time.Time) error {

	// if trying to run in the future, don't allow
	if dateTime.After(time.Now().UTC()) {
		return errors.New("cannot run cron job in the future")
	}

	// declare date types needed in query
	dateParsed := dateTime.Format("2006-01-02") // an arbitrary date must exist just to say "format kind of like this"
	date := datatypes.Date(dateTime)

	// get a connection to postgres
	dbConn := db.New()

	// start a transaction
	tx := dbConn.Begin()

	// if something goes ary, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			return
		}
	}()

	// check if we've already successfully ran this job for this date
	err := tx.Model(&db.DailyHottestCron{}).
		Where("successfully_ran = ?", date).
		First(&db.DailyHottestCron{}).
		Error
	if err != nil && err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return err
	}

	// if there already exists some entry, we've already done the job somehow
	// and we don't want to overwrite/change the results
	// this is a safety measure in case an admin accidentally tries to manually overwrite a date in the past
	var count int64
	err = tx.Model(&db.Post{}).
		Where("hottest_on = ?", dateParsed).
		Count(&count).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	if count > 0 {
		tx.Rollback()
		return nil
	}

	// manual raw SQL query to update the hottest posts for the day because Gorm is cranky
	query :=
		`WITH updated_posts AS (
        UPDATE "posts"
        SET "hottest_on" = ?, "updated_at" = ?
        WHERE "id" IN (
            SELECT "id"
            FROM "posts"
            WHERE "hidden" = false AND "hottest_on" IS NULL
            ORDER BY "trending_score" DESC
            LIMIT ?
        )
        RETURNING "school_id"
		)
		UPDATE "schools"
		SET "daily_hottests" = "daily_hottests" + (
			SELECT COUNT(*) FROM updated_posts
		)
		WHERE "id" IN (SELECT "school_id" FROM updated_posts)
		`

	// execute the raw SQL query which adds +1 to every school that has a hottest post, and updates the hottest_on date for all of the hottest posts
	err = tx.Exec(query, dateParsed, time.Now().UTC(), config.HottestPostsPageSize).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// set job on today ran as successful
	err = tx.Create(&db.DailyHottestCron{SuccessfullyRan: date}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// successfully commit transaction
	tx.Commit()
	return nil
}
