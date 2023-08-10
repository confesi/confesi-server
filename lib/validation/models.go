package validation

type CreateAccountDetails struct {
	// [required] valid email, no spaces
	Email string `json:"email" validate:"required,email,excludes= "` // intentional white space
	// [required] valid password, no spaces, at least 8 characters, at most 40 characters, must contain at least one special character
	Password string `json:"password" validate:"required,max=40,min=8,excludes= ,containsany=!@#$%^&*()_+"`
	// [optional] valid firebase token of a pre-existing account
	AlreadyExistingAccToken string `json:"already_existing_acc_token"`
}

type EmailQuery struct {
	// [required] valid email, no spaces
	Email string `json:"email" validate:"required,email,excludes= "` // intentional white space
}

type CreatePostDetails struct {
	// [required if Body empty/null] at most 100 characters
	Title string `json:"title" validate:"max=100,required_without=Body"`
	// [required if Title empty/null] at most 2000 characters
	Body string `json:"body" validate:"max=2000,required_without=Title"`
	// [required] the category of post
	Category string `json:"category" validate:"required"`
}

type CreateDraftDetails struct {
	// [required if Body empty/null] at most 100 characters
	Title string `json:"title" validate:"max=100,required_without=Body"`
	// [required if Title empty/null] at most 2000 characters
	Body string `json:"body" validate:"max=2000,required_without=Title"`
}

type SaveContentDetails struct {
	// [required] content id to save/unsave
	ContentID uint `json:"content_id" validate:"required"`
	// [required] "post" for post, "comment" for comment
	ContentType string `json:"content_type" validate:"required,oneof=post comment"`
}

type WatchedSchoolQuery struct {
	// [required] if true also return the user's home school
	IncludeHomeSchool bool `json:"include_home_school"` // not required so that the zero-value is OK
}

type SaveContentCursor struct {
	Next NullableNext `json:"next"`
}

type VoteDetail struct {
	// [required] content id to vote on
	ContentID uint `json:"content_id" validate:"required"`
	// [required] "upvote" for upvote, "downvote" for downvote
	Value *int8 `json:"value" validate:"oneof=-1 0 1"` // pointer required to "required" a zero-value (aka, vote can be 0)
	// [required] "post" for post, "comment" for comment
	ContentType string `json:"content_type" validate:"required,oneof=post comment"`
}

type PostQuery struct {
	Sort       string `json:"sort" validate:"oneof=trending new sentiment"`
	SchoolId   string `json:"school_id"`
	PurgeCache bool   `json:"purge_cache"` // true or false, doesn't have "required" so that the zero-value is OK
	SessionKey string `json:"session_key" validate:"required"`
	AllSchools bool   `json:"all_schools"` // true or false, doesn't have "required" so that the zero-value is OK
}

type WatchSchool struct {
	// [required] school id to watch
	SchoolID uint `json:"school_id" validate:"required"`
}

type UserStanding struct {
	// [required] user standing must be one of "limited", "banned", "unbanned", or "enabled"
	Standing string `json:"standing" validate:"required,oneof=limited banned enabled unbanned"`
	// [required] the user id to change standing for
	UserID string `json:"user_id" validate:"required"`
}

type UserQuery struct {
	// [required] user id to get info for
	UserID string `json:"user_id" validate:"required"`
}

type UpdateYearOfStudy struct {
	// [required] year of study to update to
	YearOfStudy string `json:"year_of_study" validate:"required"`
}

type UpdateFaculty struct {
	// [required] faculty to update to
	Faculty string `json:"faculty" validate:"required"`
}

type UpdateSchool struct {
	// [required] school to update to
	SchoolID uint `json:"school_id" validate:"required"`
}

type CreateComment struct {
	// [required] the post this comment is associated with
	PostID string `json:"post_id" validate:"required"`
	// the comment this comment is threaded under. Left empty to indicate this is a "root-level" comment
	ParentCommentID *string `json:"parent_comment_id"`
	// [required] the actual text content of the comment
	Content string `json:"content" validate:"required,min=1,max=500" gorm:"not null"`
}

type HideComment struct {
	// [required] the id of comment to "delete"
	CommentID string `json:"comment_id" validate:"required"`
}

type HidePost struct {
	// [required] the id of post to "delete"
	PostID uint `json:"post_id" validate:"required"`
}

type InitialCommentQuery struct {
	Sort       string `json:"sort" validate:"oneof=trending new"`
	PostID     string   `json:"post_id" validate:"required"`
	PurgeCache bool   `json:"purge_cache"` // true or false, doesn't have "required" so that the zero-value is OK
	SessionKey string `json:"session_key" validate:"required"`
}

type FeedbackCursor struct {
	// [required] timestamp of last seen feedback (microseconds since epoch)
	Next NullableNext `json:"next"`
}

type RepliesCommentQuery struct {
	// [required] timestamp of last seen replied comment (microseconds since epoch)
	Next NullableNext `json:"next"`
	// [required] the comment to load replies for
	ParentRoot string `json:"parent_root" validate:"required"`
}

type FeedbackDetails struct {
	// [required] feedback message
	Message string `json:"message" validate:"required,min=1,max=500"`
	// [required] feedback type
	Type string `json:"type" validate:"required"`
}

type SchoolRankQuery struct {
	// [required] school id to get rank for
	PurgeCache         bool   `json:"purge_cache"` // true or false, doesn't have "required" so that the zero-value is OK
	SessionKey         string `json:"session_key" validate:"required"`
	IncludeUsersSchool bool   `json:"include_users_school"`                // true or false, doesn't have "required" so that the zero-value is OK
	StartViewDate      string `json:"start_view_date" validate:"required"` // format: "YYYY-MM-DD"
}

type YourPostsQuery struct {
	// [required] timestamp of last viewed post content (microseconds since epoch)
	Next NullableNext `json:"next"`
}

type YourDraftsQuery struct {
	// [required] timestamp of last viewed post content (microseconds since epoch)
	Next NullableNext `json:"next"`
}

type YourCommentsQuery struct {
	// [required] timestamp of last viewed comment content (microseconds since epoch)
	Next NullableNext `json:"next"`
}

type UserCommentsQueryAdmin struct {
	// [required] timestamp of last viewed comment content (microseconds since epoch)
	Next NullableNext `json:"next"`
	// [required] user id to get comments for
	UserID string `json:"user_id" validate:"required"`
}

type FcmTokenQuery struct {
	// [required] fcm token
	Token string `json:"token" validate:"required"`
}

type HideContent struct {
	// [required] content id
	ContentID string `json:"content_id" validate:"required"`
	// [required] "post" for post, "comment" for comment
	ContentType string `json:"content_type" validate:"required,oneof=post comment"`
	// [required] true to hide, false to unhide (not having required with pointers to ensure zero-value is OK)
	Hide *bool `json:"hide"`
	// [optional] reason for hiding content
	Reason string `json:"reason"`
	// [required] mark as done with or still needs attention from mods
	ReviewedByMod bool `json:"reviewed_by_mod" validate:"required"`
}

type FetchRanCrons struct {
	// [required] type of cron to fetch
	Type string `json:"type" validate:"required,oneof=clear_expired_fcm_tokens daily_hottest all"`
	// [required] timestamp of last viewed cron job content (microseconds since epoch)
	Next NullableNext `json:"next"`
}

type FcmNotifictionPref struct {
	// true or falses, don't have "required" so that the zero-valuse are OK with pointers
	DailyHottest          *bool `json:"daily_hottest"`
	Trending              *bool `json:"trending"`
	RepliesToYourComments *bool `json:"replies_to_your_comments"`
	CommentsOnYourPosts   *bool `json:"comments_on_your_posts"`
	VotesOnYourComments   *bool `json:"votes_on_your_comments"`
	VotesOnYourPosts      *bool `json:"votes_on_your_posts"`
	QuotesOfYourPosts     *bool `json:"quotes_of_your_posts"`
}

type ReportQuery struct {
	// [required] content id to report
	ContentID uint `json:"content_id" validate:"required"`
	// [required] "post" for post, "comment" for comment
	ContentType string `json:"content_type" validate:"required,oneof=post comment"`
	// [required] report description
	Description string `json:"description" validate:"required,min=1,max=500"`
	// [required] report type
	Type string `json:"type" validate:"required"`
}

type UpdateReviewedByModQuery struct {
	// [required] content id to report
	ContentID string `json:"content_id" validate:"required"`
	// [required] "post" for post, "comment" for comment
	ContentType string `json:"content_type" validate:"required,oneof=post comment"`
	// [required] true to mark as reviewed, false to unmark as reviewed (not having required with pointers to ensure zero-value is OK)
	ReviewedByMod *bool `json:"reviewed_by_mod"`
}

type FetchReports struct {
	// [required] type of report to fetch (accepts anything because we have the options defined in the db)
	Type string `json:"type" validate:"required"`
	// [required] timestamp of last viewed report (microseconds since epoch)
	Next NullableNext `json:"next"`
}

type ReportCursor struct {
	Next NullableNext `json:"next"`
}

type HideLogCursor struct {
	Next NullableNext `json:"next"`
}

type RankedCommentsByReportsQuery struct {
	PurgeCache    bool   `json:"purge_cache"` // true or false, doesn't have "required" so that the zero-value is OK
	SessionKey    string `json:"session_key" validate:"required"`
	ReviewedByMod bool   `json:"reviewed_by_mod"` // true or false, doesn't have "required" so that the zero-value is OK
}

type RankedPostsByReportsQuery struct {
	PurgeCache    bool   `json:"purge_cache"` // true or false, doesn't have "required" so that the zero-value is OK
	SessionKey    string `json:"session_key" validate:"required"`
	ReviewedByMod bool   `json:"reviewed_by_mod"` // true or false, doesn't have "required" so that the zero-value is OK
}

type FetchReportsForCommentById struct {
	// [required] comment id
	CommentID string `json:"comment_id" validate:"required"`
	// [required] timestamp of last viewed content (ms since unix epoch)
	Next NullableNext `json:"next"`
}

type FetchReportsForPostById struct {
	// [required] post id
	PostID string `json:"post_id" validate:"required"`
	// [required] timestamp of last viewed content (ms since unix epoch)
	Next NullableNext `json:"next"`
}

type EditComment struct {
	// [required] comment id
	CommentID string `json:"comment_id" validate:"required"`
	// [required] the actual text content of the comment
	Content string `json:"content" validate:"required,min=1,max=500" gorm:"not null"`
}

type EditPost struct {
	// [required] post id
	PostID uint `json:"post_id" validate:"required"`
	// [required if Body empty/null] at most 100 characters
	Title string `json:"title" validate:"max=100,required_without=Body"`
	// [required if Title empty/null] at most 2000 characters
	Body string `json:"body" validate:"max=2000,required_without=Title"`
}

type EditDraft struct {
	// [required] post id
	DraftID uint `json:"draft_id" validate:"required"`
	// [required if Body empty/null] at most 100 characters
	Title string `json:"title" validate:"max=100,required_without=Body"`
	// [required if Title empty/null] at most 2000 characters
	Body string `json:"body" validate:"max=2000,required_without=Title"`
}

type DeleteDraft struct {
	// [required] post id
	DraftID uint `json:"draft_id" validate:"required"`
}
