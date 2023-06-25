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

type VoteDetail struct {
	// [required] content id to vote on
	ContentID uint `json:"content_id" validate:"required"`
	// [required] "upvote" for upvote, "downvote" for downvote
	Value *int8 `json:"value" validate:"oneof=-1 0 1"` // pointer required to "required" a zero-value (aka, vote can be 0)
	// [required] "post" for post, "comment" for comment
	ContentType string `json:"content_type" validate:"required,oneof=post comment"`
}

type UserStanding struct {
	// [required] user standing must be one of "limited", "banned", or "enabled"
	Standing string `json:"standing" validate:"required,oneof=limited banned enabled"`
}

type FeedbackDetails struct {
	// [required] feedback message
	Message string `json:"message" validate:"required"`
	// [required] feedback type
	Type string `json:"type" validate:"required"`
}
