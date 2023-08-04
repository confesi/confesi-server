package db

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"time"

	goaway "github.com/TwiN/go-away"
	"gopkg.in/guregu/null.v4"
	"gorm.io/datatypes"
)

type ModLevel struct {
	ID  uint   `gorm:"primaryKey" json:"id"`
	Mod string `gorm:"column:mod" json:"mod"`
}

type School struct {
	ID            uint    `gorm:"primaryKey" json:"id"`
	Name          string  `json:"name"`
	Abbr          string  `json:"abbr"`
	Lat           float32 `json:"lat"`
	Lon           float32 `json:"lon"`
	DailyHottests int     `json:"daily_hottests"`
	Domain        string  `json:"domain"`
	ImgUrl        string  `json:"img_url"`
}

type Faculty struct {
	ID      int         `gorm:"primaryKey" json:"-"`
	Faculty null.String `gorm:"column:faculty" json:"faculty"`
}

func (Faculty) TableName() string {
	return "faculties"
}

func (Post) TableName() string {
	return "posts"
}

func (School) TableName() string {
	return "schools"
}

func (FcmTopicPref) TableName() string {
	return "fcm_topic_prefs"
}

// ! Very important some fields are NOT serialized (json:"-")
type FcmTopicPref struct {
	ID                    uint   `gorm:"primaryKey" json:"-"`
	UserID                string `gorm:"column:user_id" json:"-"`
	DailyHottest          bool   `gorm:"column:daily_hottest" json:"daily_hottest"`
	Trending              bool   `gorm:"column:trending" json:"trending"`
	RepliesToYourComments bool   `gorm:"column:replies_to_your_comments" json:"replies_to_your_comments"`
	CommentsOnYourPosts   bool   `gorm:"column:comments_on_your_posts" json:"comments_on_your_posts"`
	VotesOnYourComments   bool   `gorm:"column:votes_on_your_comments" json:"votes_on_your_comments"`
	VotesOnYourPosts      bool   `gorm:"column:votes_on_your_posts" json:"votes_on_your_posts"`
	QuotesOfYourPosts     bool   `gorm:"column:quotes_of_your_posts" json:"quotes_of_your_posts"`
}

// ! Very important some fields are NOT serialized (json:"-")
type FcmToken struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    *string    `gorm:"column:user_id" json:"-"`
	Token     string     `gorm:"column:token" json:"token"`
	CreatedAt TimeMicros `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt TimeMicros `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (FcmToken) TableName() string {
	return "fcm_tokens"
}

type PostCategory struct {
	ID   uint   `gorm:"primaryKey" json:"-"`
	Name string `gorm:"column:name" json:"name"`
}

func (PostCategory) TableName() string {
	return "post_categories"
}

func (Draft) TableName() string {
	return "drafts"
}

type User struct {
	ID            string      `gorm:"primaryKey" json:"-"`
	CreatedAt     TimeMicros  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     TimeMicros  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	YearOfStudyID *uint8      `gorm:"column:year_of_study_id" json:"-"`
	YearOfStudy   YearOfStudy `gorm:"foreignKey:YearOfStudyID" json:"year_of_study"`
	FacultyID     *uint       `gorm:"column:faculty_id" json:"-"`
	Faculty       Faculty     `gorm:"foreignKey:FacultyID" json:"faculty"`
	SchoolID      uint        `gorm:"column:school_id" json:"-"`
	School        School      `gorm:"foreignKey:SchoolID" json:"school"`
	IsLimited     bool        `gorm:"is_limited" json:"-"`
}

// ! Very important some fields are NOT serialized (json:"-")
type SchoolFollow struct {
	ID        uint       `gorm:"primary_key;column:id" json:"id"`
	CreatedAt TimeMicros `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt TimeMicros `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	UserID    string     `gorm:"column:user_id" json:"-"`
	SchoolID  uint       `gorm:"column:school_id" json:"-"`
}

// ! Very important that SOME FIELDS ARE NOT EVER SERIALIZED TO PROTECT SENSATIVE DATA (json:"-")
type Post struct {
	ID            int             `gorm:"primary_key;column:id" json:"id"`
	CreatedAt     TimeMicros      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     TimeMicros      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	UserID        string          `gorm:"column:user_id" json:"-"`
	SchoolID      uint            `gorm:"column:school_id" json:"-"`
	School        School          `gorm:"foreignKey:SchoolID" json:"school"`
	FacultyID     *uint           `gorm:"column:faculty_id" json:"-"`
	Faculty       Faculty         `gorm:"foreignKey:FacultyID" json:"faculty"`
	YearOfStudyID *uint           `gorm:"column:year_of_study_id" json:"-"`
	YearOfStudy   YearOfStudy     `gorm:"foreignKey:YearOfStudyID" json:"year_of_study"`
	Title         string          `gorm:"column:title" json:"title"`
	Content       string          `gorm:"column:content" json:"content"`
	Downvote      uint            `gorm:"column:downvote" json:"downvote"`
	Upvote        uint            `gorm:"column:upvote" json:"upvote"`
	VoteScore     int             `gorm:"column:vote_score" json:"-"` // redundant to return to the user
	TrendingScore float64         `gorm:"column:trending_score" json:"trending_score"`
	HottestOn     *datatypes.Date `gorm:"column:hottest_on" json:"hottest_on"` // intentionally a pointer, so that it defaults to NULL when created and not specified (i.e. not its zero-value)
	Hidden        bool            `gorm:"column:hidden" json:"hidden"`
	ReportCount   uint            `gorm:"column:report_count" json:"-"`
	ReviewedByMod bool            `gorm:"column:reviewed_by_mod" json:"-"`
	Edited        bool            `gorm:"column:edited" json:"edited"`
	CategoryID    uint            `gorm:"column:category_id" json:"-"`
	Category      PostCategory    `gorm:"foreignKey:CategoryID" json:"category"`
}

// ! Very important that SOME FIELDS ARE NOT EVER SERIALIZED TO PROTECT SENSATIVE DATA (json:"-")
type Draft struct {
	ID        int        `gorm:"primary_key;column:id" json:"id"`
	CreatedAt TimeMicros `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt TimeMicros `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	UserID    string     `gorm:"column:user_id" json:"-"`
	Title     string     `gorm:"column:title" json:"title"`
	Content   string     `gorm:"column:content" json:"content"`
}

func (p *Post) CensorPost() Post {
	p.Title = goaway.Censor(p.Title)
	p.Content = goaway.Censor(p.Content)
	return *p
}

func (c *Comment) CensorComment() Comment {
	c.Content = goaway.Censor(c.Content)
	return *c
}

// ! Very important that SOME FIELDS ARE NOT EVER SERIALIZED TO PROTECT SENSATIVE DATA (json:"-")
type Comment struct {
	ID                        uint       `gorm:"primary_key;column:id" json:"id"`
	CreatedAt                 TimeMicros `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt                 TimeMicros `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	PostID                    uint       `gorm:"column:post_id" json:"post_id"`
	NumericalUser             *uint      `gorm:"column:numerical_user" json:"numerical_user"`                               // this is a pointer because it can be null
	NumericalReplyingUser     *uint      `gorm:"column:numerical_replying_user" json:"numerical_replying_user"`             // this is a pointer because it can be null
	NumericalUserIsOp         *bool      `gorm:"column:numerical_user_is_op" json:"numerical_user_is_op"`                   // this is a pointer because it can be null
	NumericalReplyingUserIsOp *bool      `gorm:"column:numerical_replying_user_is_op" json:"numerical_replying_user_is_op"` // this is a pointer because it can be null
	ParentRoot                *uint      `gorm:"column:parent_root" json:"parent_root"`
	ChildrenCount             uint       `gorm:"column:children_count" json:"children_count"`
	UserID                    string     `gorm:"column:user_id" json:"-"`
	Content                   string     `gorm:"column:content" json:"content"`
	Downvote                  uint       `gorm:"column:downvote" json:"downvote"`
	Upvote                    uint       `gorm:"column:upvote" json:"upvote"`
	VoteScore                 int        `gorm:"column:vote_score" json:"-"` // redundant to return to the user
	TrendingScore             float64    `gorm:"column:trending_score" json:"trending_score"`
	Hidden                    bool       `gorm:"column:hidden" json:"hidden"`
	ReportCount               uint       `gorm:"column:report_count" json:"-"`
	ReviewedByMod             bool       `gorm:"column:reviewed_by_mod" json:"-"`
	Edited                    bool       `gorm:"column:edited" json:"edited"`
}

func (c *Comment) ObscureIfHidden() Comment {
	if c.Hidden {
		c.Content = "[deleted]"
		c.NumericalReplyingUser = nil
		c.NumericalReplyingUserIsOp = nil
		c.NumericalUser = nil
		c.NumericalUserIsOp = nil
	}
	return *c
}

// This will store as a `time.Time` in the database,
// and as a unix time for json serialization
type TimeMicros struct {
	time.Time
}

func (t TimeMicros) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Time.UnixNano() / int64(time.Microsecond))
}

// Value converts the TimeMicros to a time.Time for database storage
func (t TimeMicros) Value() (driver.Value, error) {
	return t.Time, nil
}

func (t *TimeMicros) UnmarshalJSON(data []byte) error {
	var micros int64
	err := json.Unmarshal(data, &micros)
	if err != nil {
		return err
	}
	t.Time = time.Unix(0, micros*int64(time.Microsecond))
	return nil
}

func (t TimeMicros) MicroSeconds() int64 {
	epochMicros := t.Time.UnixNano() / int64(time.Microsecond)
	return epochMicros
}

func (t *TimeMicros) Scan(value interface{}) error {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		t.Time = v
	case []byte:
		var micros int64
		err := json.Unmarshal(v, &micros)
		if err != nil {
			return err
		}
		t.Time = time.Unix(0, micros*int64(time.Microsecond))
	default:
		return fmt.Errorf("unsupported scan value type: %T", value)
	}

	return nil
}

const (
	Neutralvote = 0
	Upvote      = 1
	Downvote    = -1
)

// ! Important not to serialize some fields!!
type Vote struct {
	ID        uint
	Vote      int    `db:"vote" json:"vote"`
	UserID    string `db:"user_id" json:"-"`
	PostID    *uint  `db:"post_id" gorm:"default:NULL" json:"post_id"`       // Either one of these FKs can be null, but the constraint
	CommentID *uint  `db:"comment_id" gorm:"default:NULL" json:"comment_id"` // is that exactly one of them is a valid FK
}

// ! Important not to serialize some fields!!
type SavedPost struct {
	CreatedAt TimeMicros `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt TimeMicros `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	UserID    string     `gorm:"column:user_id" json:"-"`
	PostID    uint       `gorm:"column:post_id" json:"post_id"`
}

// ! Important not to serialize some fields!!
type SavedComment struct {
	CreatedAt TimeMicros `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt TimeMicros `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	UserID    string     `gorm:"column:user_id" json:"-"`
	CommentID uint       `gorm:"column:comment_id" json:"comment_id"`
}

// ! Important not to serialize some fields!!
type Feedback struct {
	ID        int           `gorm:"primary_key;column:id" json:"id"`
	CreatedAt TimeMicros    `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UserID    string        `gorm:"column:user_id" json:"-"`
	Content   string        `gorm:"column:content" json:"content"`
	Type      *FeedbackType `gorm:"foreignKey:TypeID" json:"type,omitempty"` // Use "omitempty" here
	TypeID    uint          `gorm:"column:type_id" json:"-"`                 // references the feedback_type table
}

type ReportType struct {
	ID   int    `gorm:"primary_key;column:id" json:"-"`
	Type string `gorm:"column:type" json:"type"`
}

func (ReportType) TableName() string {
	return "report_types"
}

type FeedbackType struct {
	ID   int    `gorm:"primary_key;column:id" json:"-"`
	Type string `gorm:"column:type" json:"type"`
}

type YearOfStudy struct {
	ID   int         `gorm:"primaryKey" json:"-"`
	Name null.String `gorm:"column:name" json:"type"`
}

func (YearOfStudy) TableName() string {
	return "year_of_study"
}

func (FeedbackType) TableName() string {
	return "feedback_types"
}

type HideLog struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	CreatedAt TimeMicros `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt TimeMicros `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	PostID    *uint      `db:"post_id" gorm:"default:NULL" json:"-"`
	Post      *Post      `gorm:"foreignKey:PostID" json:"post,omitempty"` // Use "omitempty" here
	CommentID *uint      `db:"comment_id" gorm:"default:NULL" json:"-"`
	Comment   *Comment   `gorm:"foreignKey:CommentID" json:"comment,omitempty"` // Use "omitempty" here
	Reason    string     `gorm:"column:reason" json:"reason"`
	Removed   bool       `gorm:"column:removed" json:"removed"`
	UserID    string     `gorm:"column:user_id" json:"-"`
}

func (HideLog) TableName() string {
	return "hide_log"
}

// ! Important not to serialize some fields!!
type Report struct {
	ID             uint        `gorm:"primaryKey" json:"id"`
	CreatedAt      TimeMicros  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      TimeMicros  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	ReportedBy     string      `gorm:"column:reported_by" json:"-"`
	Description    string      `gorm:"column:description" json:"description"`
	TypeID         uint        `gorm:"column:type_id" json:"-"` // references the report_type table
	ReportType     *ReportType `gorm:"foreignKey:TypeID" json:"report_type"`
	Result         *string     `gorm:"column:result" json:"result"` // can be null
	HasBeenRemoved bool        `gorm:"column:has_been_removed" json:"has_been_removed"`
	PostID         *uint       `db:"post_id" gorm:"default:NULL" json:"-"`
	Post           *Post       `gorm:"foreignKey:PostID" json:"post,omitempty"` // Use "omitempty" here
	CommentID      *uint       `db:"comment_id" gorm:"default:NULL" json:"-"`
	Comment        *Comment    `gorm:"foreignKey:CommentID" json:"comment,omitempty"` // Use "omitempty" here
}

type CronJob struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt TimeMicros     `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	Ran       datatypes.Date `gorm:"column:ran" json:"ran"`
	Type      string         `gorm:"column:type" json:"type"`
}

func (CronJob) TableName() string {
	return "cron_jobs"
}

func (Report) TableName() string {
	return "reports"
}
