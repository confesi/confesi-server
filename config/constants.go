package config

const (
	MaxCommentThreadDepthExcludingRoot = 1              // how many levels of comments are allowed (depth)
	RootCommentsLoadedInitially        = 5              // how many root comments are returned at once
	RepliesLoadedInitially             = 5              // how many replies are returned at once
	RepliesLoadedManually              = 5              // how many replies are loaded when user clicks on "load more replies"
	HottestPostsPageSize               = 5              // how many hottest posts are returned in a page size per day
	DailyHottestCronJobResultsPageSize = 10             // how many cron job results are returned in a page size
	FeedPostsPageSize                  = 10             // how many posts are returned in a page size in the regular feed
	RankedSchoolsPageSize              = 10             // how many schools are returned in a page size in the ranked schools page
	RedisPostsCache                    = "posts"        // redis key for posts cache
	RedisSchoolsRankCache              = "schools_rank" // redis key for schools rank cache
	RedisCommentsCache                 = "comments"     // redis key for comments cache
)

// Other constants are defined directly in the request struct validation tags, here: ~/lib/validation/models.go
