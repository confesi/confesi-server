package config

const (
	MaxCommentThreadDepthExcludingRoot = 1  // how many levels of comments are allowed (depth)
	RootCommentsLoadedInitially        = 2  // how many root comments are returned at once
	RepliesLoadedInitially             = 2  // how many replies are returned at once
	RepliesLoadedManually              = 2  // how many replies are loaded when user clicks on "load more replies"
	HottestPostsPageSize               = 5  // how many hottest posts are returned in a page size per day
	DailyHottestCronJobResultsPageSize = 10 // how many cron job results are returned in a page size
	FeedPostsPageSize                  = 10 // how many posts are returned in a page size in the regular feed
)
