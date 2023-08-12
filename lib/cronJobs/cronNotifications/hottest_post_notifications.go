package cronNotifications

import (
	"confesi/config"
	"confesi/config/builders"
	"confesi/db"
	"confesi/lib/cronJobs"
	"confesi/lib/fire"
	"confesi/lib/logger"
	"errors"
	"time"

	fcmMsg "confesi/lib/firebase_cloud_messaging"

	"github.com/go-co-op/gocron"

	"gorm.io/datatypes"
)

// Cron job that runs every  two hours to send notifications to users about the hottest posts.
func StartDailyHottestPostsCronJob() {

	upperBound, err := time.Parse("15:04", config.HottestPostNotificationsUpperBound)
	if err != nil {
		panic(err)
	}
	lowerBound, err := time.Parse("15:04", config.HottestPostNotificationsLowerBound)
	if err != nil {
		panic(err)
	}
	intervalTime := upperBound.Sub(lowerBound)
	interval := intervalTime.Hours()

	s := gocron.NewScheduler(time.UTC)
	s.Every(interval).Hour().Do(func() {
		cronJobs.RetryLoop(1000, 1000*60, 2.0, 2, func() error {
			return DoHottestPostNotifications(time.Now().UTC())
		})
	})
	logger.StdInfo("started scheduler for daily hottest notification cron job")
	s.StartAsync()
}

func DoHottestPostNotifications(dateTime time.Time) error {

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

	// obtain the hottest posts by date, order by hottest_on desc from posts table, but
	var hottestPosts []db.Post
	err := tx.Model(&db.Post{}).
		Order("hottest_on desc").
		Where("hottest_on IS NOT NULL").
		Limit(config.HottestPostsPageSize).
		Find(&hottestPosts).
		Error

	if err != nil {
		tx.Rollback()
		return err
	}

	// obtain the school ids of the hottest posts
	var hottestPostSchoolIds []uint
	for _, post := range hottestPosts {
		hottestPostSchoolIds = append(hottestPostSchoolIds, post.SchoolID.Val)
	}

	occurences := occurences(hottestPostSchoolIds)

	// get the schools from the database off the school ids
	var schools []db.School
	err = tx.Model(&db.School{}).
		Where("id IN ?", hottestPostSchoolIds).
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

		// if school timezone is between 10 am and 1 pm, send notifications
		if schoolTimeZoneParsed >= config.HottestPostNotificationsLowerBound && schoolTimeZoneParsed <= config.HottestPostNotificationsUpperBound {
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

			tokens := []string{}

			//Obtain fcm_tokens.tokens from fcm_tokens table where user_id is in users table and school_id is in schools table
			err = tx.
				Table("fcm_tokens").
				Select("fcm_tokens.token").
				Joins("JOIN users ON users.id = fcm_tokens.user_id").
				Joins("JOIN schools ON schools.id = users.school_id").
				Where("schools.id = ?", school.ID.Val).
				Pluck("fcm_tokens.token", &tokens).
				Error

			if err != nil {
				tx.Rollback()
				return err
			}

			// send notifications to users

			hottestOccurences := occurences[school.ID.Val]

			go fcmMsg.New(msgClient).
				ToTokens(tokens).
				WithMsg(builders.YourSchoolsDailyHottestNoti(hottestOccurences)).
				WithData(builders.YourSchoolsDailyHottestData()).
				Send(*tx)

		}
	}

	// set job on today ran as successful
	err = tx.Create(&db.CronJob{Ran: date, Type: cronJobs.HottestPostNotificationsCronJob}).Error
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

// Create a function that returns a dictionary of the number of occurences of each item in a list
func occurences(list []uint) map[uint]int {
	// Create a dictionary
	dict := make(map[uint]int)
	// Iterate through the list
	for _, item := range list {
		// Check if the item is in the dictionary
		_, exist := dict[item]
		// If it is, increment the value
		if exist {
			dict[item] += 1
		} else {
			// If it's not, set the value to 1
			dict[item] = 1
		}
	}
	return dict
}
