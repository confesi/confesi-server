package validation

type CreateAccountDetails struct {
	// [required] valid email, no spaces
	Email string `json:"email" validate:"required,email,excludes= "`
	// [required] valid password, no spaces, at least 8 characters, at most 40 characters, must contain at least one special character
	Password string `json:"password" validate:"required,max=40,min=8,excludes= ,containsany=!@#$%^&*()_+"`
	// [required] year of study, must be between 1 and 8 (inclusive)
	YearOfStudy uint8 `json:"year_of_study" validate:"required,gte=1,lte=8"`
	// [required] we'll do validation later against the postgres table
	Faculty string `json:"faculty" validate:"required"`
}

type CreatePostDetails struct {
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

type SaveContentCursor struct {
	// [required] timestamp of last saved content (ms since epoch)
	Next uint `json:"next" validate:"required"`
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
	Sort       string `json:"sort" validate:"oneof=trending new"`
	School     uint   `json:"school" validate:"required"`
	PurgeCache bool   `json:"purge_cache"` // true or false, doesn't have "required" so that the zero-value is OK
	SessionKey string `json:"session_key" validate:"required"`
}

type WatchSchool struct {
	// [required] school id to watch
	SchoolID uint `json:"school_id" validate:"required"`
}

type UserStanding struct {
	// [required] user standing must be one of "limited", "banned", or "enabled"
	Standing string `json:"standing" validate:"required,oneof=limited banned enabled"`
}

type CreateComment struct {
	// [required] the post this comment is associated with
	PostID uint `json:"post_id" validate:"required"`
	// the comment this comment is threaded under. Left empty to indicate this is a "root-level" comment
	ParentCommentID *int64 `json:"parent_comment_id"`
	// [required] the actual text content of the comment
	Content string `json:"content" validate:"required,min=1,max=500" gorm:"not null"`
}

type HideComment struct {
	// [required] the id of comment to delete
	CommentID uint `json:"comment_id" validate:"required"`
}

type InitialCommentQuery struct {
	Sort       string `json:"sort" validate:"oneof=trending new"`
	PostID     uint   `json:"post_id" validate:"required"`
	PurgeCache bool   `json:"purge_cache"` // true or false, doesn't have "required" so that the zero-value is OK
	SessionKey string `json:"session_key" validate:"required"`
}

type RepliesCommentQuery struct {
	// [required] timestamp of last seen replied comment (ms since epoch)
	Next uint `json:"next" validate:"required"`
	// [required] the comment to load replies for
	ParentComment uint `json:"parent_comment" validate:"required"`
}

type FeedbackDetails struct {
	// [required] feedback message
	Message string `json:"message" validate:"required"`
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
