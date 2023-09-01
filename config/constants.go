package config

// ----- APPLICATION CONSTANTS

const (
	MaxCommentThreadDepthExcludingRoot               = 1                                        // how many levels of comments are allowed (depth)
	RootCommentsLoadedInitially                      = 10                                       // how many root comments are returned at once
	RepliesLoadedInitially                           = 2                                        // how many replies are returned at once
	RepliesLoadedManually                            = 5                                        // how many replies are loaded when user clicks on "load more replies"
	HottestPostsPageSize                             = 5                                        // how many hottest posts are returned in a page size per day
	CronJobPageSize                                  = 10                                       // how many cron job results are returned in a page size
	ReportsPageSize                                  = 10                                       // how many reports are returned in a page size
	FeedPostsPageSize                                = 10                                       // how many posts are returned in a page size in the regular feed
	RankedSchoolsPageSize                            = 2                                        // how many schools are returned in a page size in the ranked schools page
	RedisPostsCache                                  = "posts"                                  // redis key for posts cache
	RedisRoomsCache                                  = "rooms"                                  // redis key for rooms cache
	RedisSchoolsRankCache                            = "schools_rank"                           // redis key for schools rank cache
	RedisCommentsCache                               = "comments"                               // redis key for comments cache
	RedisCommentsCacheByReports                      = "comments_by_report"                     // redis key for comments cache by reports
	RedisPostsCacheByReports                         = "posts_by_report"                        // redis key for posts cache by reports
	RedisRateLimitingCache                           = "ip_rate_limiting"                       // redis key for ip-based rate limiting cache
	RedisEmailRateLimitingRouteKeyResendVerification = "email_rate_limiting_email_verification" // redis key for email rate limiting cache
	RedisEmailRateLimitingRouteKeySendPwReset        = "email_rate_limiting_pw_reset"           // redis key for email rate limiting cache
	RedisGlobalUserStats                             = "global_user_stats"                      // redis key for global user stats cache
	YourPostsPageSize                                = 3                                        // how many posts are returned in a page size in the your posts page
	YourDraftsPageSize                               = 3                                        // how many drafts are returned in a page size in the your drafts page
	YourCommentsPageSize                             = 3                                        // how many comments are returned in a page size in the your comments page
	SavedPostsAndCommentsPageSize                    = 2                                        // how many posts and comments are returned in a page size in the saved posts and comments page
	UserCommentsAdminPageSize                        = 3                                        // how many comments are returned in a page size in the user comments admin page
	ViewYourReportsPageSize                          = 10                                       // how many reports are returned in a page size in the view your reports page
	AdminCommentsSortedByReportsPageSize             = 10                                       // how many comments are returned in a page size in the admin comments sorted by reports page
	AdminPostsSortedByReportsPageSize                = 10                                       // how many posts are returned in a page size in the admin posts sorted by reports page
	AdminViewAllReportsPerCommentId                  = 10                                       // how many reports are returned in a page size in the view of reports for a comment given by id
	AdminViewAllReportsPerPostId                     = 10                                       // how many reports are returned in a page size in the view of reports for a post given by id
	AdminFeedbackPageSize                            = 10                                       // how many feedbacks are returned in a page size in the admin feedback page
	QueryForSchoolsBySearchPageSize                  = 10                                       // how many schools are returned in a page size in the search schools page
	QueryForSchoolsBySearchFloorSimilarityMatchValue = 0.05                                     // how much of a match is required for a school to be returned in the search schools page (value is the min percent of match, IE: 0.05 = 5% match at least, or 1 in 20 of the characters match)
	DefaultRange                                     = 50000                                    // default distance to include schools in, in meters
	HottestPostNotificationsLowerBound               = "00:00"                                  // the lower bound of the hottest post notifications cron job - 24h time
	HottestPostNotificationsUpperBound               = "23:59"                                  // the upper bound of the hottest post notifications cron job - 24h time
	TitleMaxLength                                   = 100                                      // max length of a post title
	BodyMaxLength                                    = 1000                                     // max length of a post content/body
)

// ----- DEPLOYMENT CONSTANTS

const (
	iOSBundleId        = "com.confesi.app"
	AndroidPackageName = "com.confesi.app"
	FirebaseProjectID  = "confesi-server-dev" // Firebase project id
	Development        = true                 // development mode
)

// ---- AWS CONSTANTS

const (
	AwsUserUploadsBucket              string  = "confesi-uploads"
	AwsRekognitionConfidenceThreshold float64 = 80
	AwsUserUploadsBucketBaseUrl       string  = "https://confesi-uploads.s3.us-east-2.amazonaws.com/"
)

// ----- ALT CONSTANTS

// -> pther constants are defined directly in the request struct validation tags, here: ~/lib/validation/models.go
