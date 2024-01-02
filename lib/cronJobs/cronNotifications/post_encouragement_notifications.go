package cronNotifications

import (
	"confesi/config"
	"confesi/config/builders"
	"confesi/db"
	"confesi/lib/cronJobs"
	"confesi/lib/fire"
	fcmMsg "confesi/lib/firebase_cloud_messaging"
	"confesi/lib/logger"
	"errors"
	"time"

	"github.com/go-co-op/gocron"
	"gorm.io/datatypes"
)

// Cron job that runs every  two hours to send notifications to users about the hottest posts.
func StartPostEncouragementCronJob() {
	timesPerDay := config.PostEncouragementNotificationsTimesPerDay

	// Obtain the interval between each run of the cron job
	interval := time.Duration(24/timesPerDay) * time.Hour

	// Create a cron job scheduler
	s := gocron.NewScheduler(time.UTC)

	// Run the cron job every interval hours
	s.Every(interval).Hour().Do(func() {
		cronJobs.RetryLoop(1000, 1000*60, 2.0, 1, func() error {
			return DoPostEncouragementNotifications(time.Now().UTC())
		})
	})
	// Log that the cron job has started
	logger.StdInfo("started scheduler for post encouragement cron job")
	s.StartAsync()
}

func DoPostEncouragementNotifications(dateTime time.Time) error {

	// if trying to run in the future, don't allow
	if dateTime.After(time.Now().UTC()) {
		return errors.New("cannot run cron job in the future")
	}

	// declare date types needed in query

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

	// get the schools from the database
	var schools []db.School
	err := tx.Model(&db.School{}).
		Find(&schools).
		Error

	if err != nil {
		tx.Rollback()
		return err
	}

	// Obtain Message Client
	firebaseInstance := fire.New()
	msgClient := firebaseInstance.MsgClient

	// send notifications to users if school timezone is between 10 am and 1 pm
	for _, school := range schools {
		// convert school timezone to time.Time
		schoolTimezoneLocation, err := time.LoadLocation(school.Timezone)
		if err != nil {
			tx.Rollback()
			return err
		}
		schoolTimeZoneParsed := time.Now().In(schoolTimezoneLocation).Format("15:04")

		// if school timezone is between 10 am and 7 pm, send notifications
		if schoolTimeZoneParsed >= config.PostEncouragementNotificationsLowerBound && schoolTimeZoneParsed <= config.PostEncouragementNotificationsUpperBound {
			// get the users of the school
			var users []db.User
			err = tx.Model(&db.User{}).
				Where("school_id = ?", school.ID.Val).
				Find(&users).
				Error

			if err != nil {
				tx.Rollback()
				return err
			}
			// Obtain the date one day ago
			timeInPast := time.Now().UTC().AddDate(0, 0, -config.PostEncouragementNotificationsDaysWithoutNotifications)

			// Obtain Users who have not been notified yet today
			var usersNotifiedAlready []db.User
			err = tx.Table("notification_logs").
				Select("user_id").
				Where("created_at > ?", timeInPast).
				Where("user_id IN ?", users).
				Pluck("user_id", &usersNotifiedAlready).
				Error

			if err != nil {
				tx.Rollback()
				return err
			}

			// TODO: Add randomization, grabbing x users at a time (May be inefficient)
			// Obtain Users who have not been notified yet today
			var usersNotNotifiedYet []db.User
			err = tx.Model(&db.User{}).
				Where("id NOT IN ?", usersNotifiedAlready).
				Find(&usersNotNotifiedYet).
				Error

			if err != nil {
				tx.Rollback()
				return err
			}

			tokens := []string{}

			//Obtain fcm_tokens.tokens from fcm_tokens table where user_id is in users table and school_id is in schools table
			err = tx.
				Table("fcm_tokens").
				Select("fcm_tokens.token").
				Where("user_id = ?", usersNotNotifiedYet).
				Pluck("fcm_tokens.token", &tokens).
				Error

			if err != nil {
				tx.Rollback()
				return err
			}

			// send notifications to users

			go fcmMsg.New(msgClient).
				ToTokens(tokens).
				WithMsg(builders.PostEncouragementNoti()).
				WithData(builders.PostEncouragementData()).
				Send()

		}
	}

	// set job on today ran as successful
	err = tx.Create(&db.CronJob{Ran: date, Type: cronJobs.PostEncouragementNotificationsCronJob}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
