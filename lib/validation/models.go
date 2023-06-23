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
