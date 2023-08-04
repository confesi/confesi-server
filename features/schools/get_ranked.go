package schools

import (
	"confesi/config"
	"confesi/lib/logger"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

const (
	seenSchoolsCacheExpiry = 24 * time.Hour // one day
)

type rankedSchoolsResult struct {
	Schools    []SchoolDetail `json:"schools"`
	UserSchool *SchoolDetail  `json:"user_school"`
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

	type doesExist struct {
		Exists bool
	}

	var result doesExist

	err = tx.Raw(`
		SELECT EXISTS (
			SELECT 1
			FROM cron_jobs
			WHERE cron_jobs.ran = ?
			AND type = 'daily_hottest'
		) AS exists
	`, nextDate).Scan(&result).Error

	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// if we found some records, then the cron job has already run for the next day
	found := result.Exists

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

	var possibleRestriction string
	if len(ids) > 0 {
		idsStr := strings.Join(ids, ", ") // Convert the ids slice to a comma-separated string
		possibleRestriction = "WHERE s.id NOT IN (" + idsStr + ")"
		fmt.Println(possibleRestriction)
	}

	query := h.DB.Raw(`
		SELECT s.*, 
			COALESCE(u.school_id = s.id, false) as home,
			CASE 
				WHEN EXISTS (SELECT 1 FROM school_follows WHERE user_id = ? AND school_id = s.id)
				THEN true
				ELSE false
			END as watched
		FROM schools as s
		LEFT JOIN (
			SELECT DISTINCT school_id
			FROM users
			WHERE id = ?
		) as u ON u.school_id = s.id
		`+possibleRestriction+`
		GROUP BY s.id, u.school_id
		ORDER BY s.daily_hottests DESC
		LIMIT ?;
	`, token.UID, token.UID, config.RankedSchoolsPageSize)

	err = query.Find(&schoolResult.Schools).Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// retrieve the user's school if desired, but don't add to cache!
	if req.IncludeUsersSchool {
		err := h.DB.Raw(`
		SELECT s.*, 
			COALESCE(u.school_id = s.id, false) as home,
			CASE 
				WHEN EXISTS (SELECT 1 FROM school_follows WHERE user_id = ? AND school_id = s.id)
				THEN true
				ELSE false
			END as watched
		FROM schools as s
		LEFT JOIN (
			SELECT DISTINCT school_id
			FROM users
			WHERE id = ?
		) as u ON u.school_id = s.id
		JOIN users ON s.id = users.school_id
		WHERE users.school_id = s.id
		LIMIT 1;
	`, token.UID, token.UID).Scan(&schoolResult.UserSchool).Error

		if err != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}

		latlong, err := utils.GetLatLong(c)
		if err == nil {
			coord := Coordinate{lat: latlong.Lat, lon: latlong.Long, radius: config.DefaultRange}
			distance := coord.getDistance(schoolResult.UserSchool.School)
			schoolResult.UserSchool.Distance = &distance
		}
	}

	addLatLong := false
	latlong, err := utils.GetLatLong(c)
	if err == nil {
		addLatLong = true
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
		if addLatLong {
			coord := Coordinate{lat: latlong.Lat, lon: latlong.Long, radius: config.DefaultRange}
			school := &schoolResult.Schools[i]
			distance := coord.getDistance(school.School)
			school.Distance = &distance
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
