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

// TODO: update length of title and body, and other stuff
type CreatePostDetails struct {
	Title string `json:"title" validate:"required,max=100"`
	Body  string `json:"body" validate:"required,max=2000"`
}
