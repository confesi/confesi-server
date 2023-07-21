package docs

import "time"

type Success struct {
	Error *string `json:"error" example:"null"`
	Value *string `json:"value" example:"null"`
}

type Created struct {
	Error *string `json:"error" example:"null"`
	Value *string `json:"value" example:"null"`
}

type ServerError struct {
	Error *string `json:"error" example:"server error"`
	Value *string `json:"value" example:"null"`
}

type InvalidDateFormat struct {
	Error *string `json:"error" example:"invalid date format"`
	Value *string `json:"value" example:"null"`
}

type HottestPosts struct {
	Error *string `json:"error" example:"null"`
	Value *string `json:"value" example:"null"`
}

type PostNotFound struct {
	Error *string `json:"error" example:"post not found"`
	Value *string `json:"value" example:"null"`
}

type PostRemoved struct {
	Error *string `json:"error" example:"post removed"`
	Value *string `json:"value" example:"null"`
}

type InvalidSortField struct {
	Error *string `json:"error" example:"invalid sort field"`
	Value *string `json:"value" example:"null"`
}

type FailedToUpdateCache struct {
	Error *string `json:"error" example:"failed to update cache"`
	Value *string `json:"value" example:"null"`
}

type FailedToSetCacheExpiration struct {
	Error *string `json:"error" example:"failed to set cache expiration"`
	Value *string `json:"value" example:"null"`
}

type CachePurged struct {
	Error *string `json:"error" example:"null"`
	Value *string `json:"value" example:"null"`
}

type NoPostFoundWithThisID struct {
	Error *string `json:"error" example:"no post found with this ID"`
	Value *string `json:"value" example:"null"`
}

type PostHidden struct {
	Error *string `json:"error" example:"null"`
	Value *string `json:"value" example:"null"`
}

type SentimentAnaylsis struct {
	Error *string `json:"error" example:"null"`
	Value struct {
		Positive int `json:"positive" example:"0"`
		Negative int `json:"negative" example:"0"`
		Neutral  int `json:"neutral" example:"1"`
		Compound int `json:"compound" example:"0"`
	} `json:"value"`
}

type GetCronJobs struct {
	Error *string `json:"error" example:"null"`
	Value struct {
		Crons []struct {
			ID        int       `json:"id" example:"1"`
			CreatedAt int64     `json:"created_at" example:"1689736231093838"`
			Ran       time.Time `json:"ran" example:"2023-06-21T00:00:00Z"`
			Type      string    `json:"type" example:"clear_expired_fcm_tokens"`
		} `json:"crons"`
		Next int64 `json:"next" example:"1689736231093838"`
	} `json:"value"`
}

type FeedbackByID struct {
	Error *string `json:"error" example:"null"`
	Value struct {
		ID        int    `json:"id" example:"12"`
		CreatedAt int64  `json:"created_at" example:"1687339205605989"`
		Content   string `json:"content" example:"WOAH THIS WAS SUCH AN AMAZING APP I CAN'T BELEIVE YOU DID IT!"`
		TypeID    int    `json:"type_id" example:"0"`
	} `json:"value"`
}

type RegisteredUser struct {
	Error *string `json:"error" example:"null"`
	Value string  `json:"value" example:"auth"`
}

type ErrorExtractingDomainFromEmail struct {
	Error *string `json:"error" example:"error extracting domain from email"`
	Value string  `json:"value" example:"null"`
}

type DomainDoesNotBelongToASchool struct {
	Error *string `json:"error" example:"domain doesn't belong to school"`
	Value string  `json:"value" example:"null"`
}

type FacultyDoesNotExist struct {
	Error *string `json:"error" example:"faculty doesn't exist"`
	Value string  `json:"value" example:"null"`
}

type EmailAlreadyExists struct {
	Error *string `json:"error" example:"email already exists"`
	Value string  `json:"value" example:"null"`
}

type ParentCommentDoesNotExist struct {
	Error *string `json:"error" example:"parent-comment and post combo doesn't exist"`
	Value string  `json:"value" example:"null"`
}

type PostFound struct {
	Error *string `json:"error" example:"null"`
	Value struct {
		Post struct {
			ID        int   `json:"id" example:"4"`
			CreatedAt int64 `json:"created_at" example:"1689570882898185"`
			UpdatedAt int64 `json:"updated_at" example:"1689570882898185"`
			School    struct {
				Name          string  `json:"name" example:"University of Victoria"`
				Abbr          string  `json:"abbr" example:"UVIC"`
				Lat           float64 `json:"lat" example:"48.4634"`
				Lon           float64 `json:"lon" example:"-123.3117"`
				DailyHottests int     `json:"daily_hottests" example:"0"`
				Domain        string  `json:"domain" example:"@uvic.ca"`
			} `json:"school"`
			Faculty struct {
				Faculty string `json:"faculty" example:"ENG"`
			} `json:"faculty"`
			Title         string  `json:"title" example:"noti"`
			Content       string  `json:"content" example:""`
			Downvote      int     `json:"downvote" example:"0"`
			Upvote        int     `json:"upvote" example:"0"`
			TrendingScore int     `json:"trending_score" example:"0"`
			HottestOn     *string `json:"hottest_on" example:"null"`
		} `json:"post"`
		UserVote int       `json:"user_vote" example:"0"`
		Owner    bool      `json:"owner" example:"true"`
		Emojis   []*string `json:"emojis" example:"[]"`
	} `json:"value"`
}

type YourPosts struct {
	Error *string `json:"error" example:"null"`
	Value struct {
		Posts []struct {
			Post struct {
				ID        int   `json:"id" example:"1"`
				CreatedAt int64 `json:"created_at" example:"1689559373125623"`
				UpdatedAt int64 `json:"updated_at" example:"1689559373125623"`
				School    struct {
					Name          string  `json:"name" example:"University of Victoria"`
					Abbr          string  `json:"abbr" example:"UVIC"`
					Lat           float64 `json:"lat" example:"48.4634"`
					Lon           float64 `json:"lon" example:"-123.3117"`
					DailyHottests int     `json:"daily_hottests" example:"0"`
					Domain        string  `json:"domain" example:"@uvic.ca"`
				} `json:"school"`
				Faculty struct {
					Faculty string `json:"faculty" example:"ENG"`
				} `json:"faculty"`
				Title         string  `json:"title" example:"noti"`
				Content       string  `json:"content" example:"noti"`
				Downvote      int     `json:"downvote" example:"0"`
				Upvote        int     `json:"upvote" example:"0"`
				TrendingScore int     `json:"trending_score" example:"0"`
				HottestOn     *string `json:"hottest_on" example:"null"`
			} `json:"post"`
			UserVote int       `json:"user_vote" example:"0"`
			Owner    bool      `json:"owner" example:"true"`
			Emojis   []*string `json:"emojis" example:"[]"`
		} `json:"posts"`
		Next int64 `json:"next" example:"1689641897327536"`
	} `json:"value"`
}
