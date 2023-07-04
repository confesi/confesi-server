package schools

import (
	"confesi/config"
	"confesi/db"
	"confesi/lib/logger"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

const (
	seenSchoolsCacheExpiry = 24 * time.Hour // one day
)

type rankedSchoolsResult struct {
	Schools    []db.School `json:"schools"`
	UserSchool *db.School  `json:"user_school"`
}

func (h *handler) handleGetRankedSchools(c *gin.Context) {
	// extract request
	var req validation.SchoolRankQuery
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// Parse the date string into a time.Time value
	date, err := time.Parse("2006-01-02", req.StartViewDate) // this basically says YYYY-MM-DD, not sure why, but it only works with a dummy date example?
	nextDate := date.AddDate(0, 0, 1)
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid date format").Send(c)
		return
	}

	// start a transaction
	tx := h.DB.Begin()

	// if something goes ary, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}()

	var found bool

	err = tx.Raw(`
		SELECT EXISTS (
			SELECT 1
			FROM daily_hottest_cron_jobs
			WHERE daily_hottest_cron_jobs.successfully_ran = ?
		)
	`, nextDate).
		Scan(&found).
		Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// session key that can only be created by *this* user, so it can't be guessed to manipulate others' feeds
	idSessionKey, err := utils.CreateCacheKey(config.RedisSchoolsRankCache, token.UID, req.SessionKey)
	if err != nil {
		response.New(http.StatusBadRequest).Err(utils.UuidError.Error()).Send(c)
		return
	}

	// if found, it means the cron job has already run for the next day meaning our data is now invalid
	// so we need to return an error and clear the user's seen id cache
	if req.PurgeCache || found {
		// purge the cache
		err := h.redis.Del(c, idSessionKey).Err()
		if err != nil {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}
	if found {
		response.New(http.StatusGone).Err("data is invalid, please refresh").Send(c)
		return
	}

	// retrieve the school IDs from the cache
	ids, err := h.redis.SMembers(c, idSessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			ids = []string{} // assigns an empty slice
		} else {
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	schoolResult := rankedSchoolsResult{}

	query := tx.
		Order("daily_hottests DESC").
		Limit(config.RankedSchoolsPageSize)

	if len(ids) > 0 {
		query = query.Where("schools.id NOT IN (?)", ids)
	}

	err = query.Find(&schoolResult.Schools).Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// retrieve the user's school if desired, but don't add to cache!
	if req.IncludeUsersSchool {
		err := tx.
			Table("schools").
			Joins("JOIN users ON schools.id = users.school_id").
			Where("users.school_id = schools.id"). // redundant `where` clause?
			First(&schoolResult.UserSchool).
			Error
		if err != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	// update the cache with the retrieved schools IDs
	for i := range schoolResult.Schools {
		id := fmt.Sprint(schoolResult.Schools[i].ID)
		err := h.redis.SAdd(c, idSessionKey, id).Err()
		if err != nil {
			logger.StdErr(err)
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err("failed to update cache").Send(c)
			return
		}
	}

	// set the expiration for the cache
	err = h.redis.Expire(c, idSessionKey, seenSchoolsCacheExpiry).Err()
	if err != nil {
		logger.StdErr(err)
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err("failed to set cache expiration").Send(c)
		return
	}

	// commit the transaction
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// if all good, send 200
	response.New(http.StatusOK).Val(schoolResult).Send(c)
}
