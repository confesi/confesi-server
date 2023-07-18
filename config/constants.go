package config

const (
	MaxCommentThreadDepthExcludingRoot = 1                     // how many levels of comments are allowed (depth)
	RootCommentsLoadedInitially        = 5                     // how many root comments are returned at once
	RepliesLoadedInitially             = 1                     // how many replies are returned at once
	RepliesLoadedManually              = 5                     // how many replies are loaded when user clicks on "load more replies"
	HottestPostsPageSize               = 5                     // how many hottest posts are returned in a page size per day
	CronJobPageSize                    = 10                    // how many cron job results are returned in a page size
	ReportsPageSize                    = 10                    // how many reports are returned in a page size
	FeedPostsPageSize                  = 10                    // how many posts are returned in a page size in the regular feed
	RankedSchoolsPageSize              = 10                    // how many schools are returned in a page size in the ranked schools page
	RedisPostsCache                    = "posts"               // redis key for posts cache
	RedisSchoolsRankCache              = "schools_rank"        // redis key for schools rank cache
	RedisCommentsCache                 = "comments"            // redis key for comments cache
	RedisCommentsCacheByReports        = "comments_by_reports" // redis key for comments cache by reports
	RedisRateLimitingCache             = "rate_limiting"       // redis key for rate limiting cache
	YourPostsPageSize                  = 3                     // how many posts are returned in a page size in the your posts page
	YourCommentsPageSize               = 3                     // how many comments are returned in a page size in the your comments page
	SavedPostsAndCommentsPageSize      = 2                     // how many posts and comments are returned in a page size in the saved posts and comments page
	UserCommentsAdminPageSize          = 3                     // how many comments are returned in a page size in the user comments admin page
	ViewYourReportsPageSize            = 10                    // how many reports are returned in a page size in the view your reports page
)

// Other constants are defined directly in the request struct validation tags, here: ~/lib/validation/models.go
