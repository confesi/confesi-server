package dailyHottestPosts

import (
	"confesi/config"
	"confesi/config/builders"
	"confesi/db"
	"confesi/lib/awards"
	"confesi/lib/cache"
	"confesi/lib/cronJobs"
	"confesi/lib/fire"
	"confesi/lib/logger"
	"confesi/lib/utils"
	"context"
	"errors"
	"time"

	fcm "confesi/lib/firebase_cloud_messaging"

	"github.com/go-co-op/gocron"
	"gorm.io/datatypes"

	"gorm.io/gorm"
)

// Cron job that runs daily to update the hottest posts.
func StartDailyHottestPostsCronJob(fb *fire.FirebaseApp) {
	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Day().At(config.WhenRunDailyHottestCron).Do(func() {
		cronJobs.RetryLoop(1000, 1000*60, 6.0, 20, func() error {
			return DoDailyHottestJob(time.Now().UTC(), fb)
		})
	})
	logger.StdInfo("started scheduler for daily hottest posts cron job")
	s.StartAsync()
}

func DoDailyHottestJob(dateTime time.Time, fb *fire.FirebaseApp) error {

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
	err := tx.Model(&db.CronJob{}).
		Where("ran = ?", date).
		Where("type = ?", cronJobs.DailyHottestCronJobLog).
		First(&db.CronJob{}).
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

	var postIDs []db.EncryptedID

	err = tx.Model(&db.Post{}).Select("id").Where("hidden = false AND hottest_on IS NULL").Order("trending_score DESC").Limit(config.HottestPostsPageSize).Scan(&postIDs).Error
	if err != nil {
		// handle error
		tx.Rollback()
		return err
	}

	// Now you can pass the postIDs to the awards function
	err = awards.OnPostBecomingHottest(tx, postIDs)
	if err != nil {
		// handle error
		tx.Rollback()
		return err
	}

	// execute the raw SQL query which adds +1 to every school that has a hottest post, and updates the hottest_on date for all of the hottest posts
	err = tx.Exec(query, dateParsed, time.Now().UTC(), config.HottestPostsPageSize).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// set job on today ran as successful
	err = tx.Create(&db.CronJob{Ran: date, Type: cronJobs.DailyHottestCronJobLog}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// clear the user's seen id cache
	folder_name := config.RedisSchoolsRankCache
	store := cache.New() //Redis client
	c := context.TODO()  // context

	err = utils.DeleteCacheFolder(&c, store, folder_name)
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

	for _, postID := range postIDs {
		var tokens []string

		// Query for FCM tokens of the user who owns the post
		err := dbConn.Table("fcm_tokens").
			Select("fcm_tokens.token").
			Joins("JOIN users ON users.id = fcm_tokens.user_id").
			Joins("JOIN posts ON posts.user_id = users.id").
			Pluck("fcm_tokens.token", &tokens).
			Error

		// Send notifications if tokens are found (don't error-out if FCM doesn't work!)
		if err != nil {
			return err
		}

		if err == nil && len(tokens) > 0 {
			go fcm.New(fb.MsgClient).
				ToTokens(tokens).
				WithMsg(builders.YouReachedDailyHottestNoti()).
				WithData(builders.YouReachedDailyHottestData(postID.ToMasked())).
				Send()
		}
	}

	return nil
}
