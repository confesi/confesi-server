package db

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"

	"time"

	"github.com/lib/pq"
)

const (
	ModEnableID = 1
	ModEnable   = "enabled"

	ModLimitedID = 2
	ModLimited   = "limited"

	ModBannedID = 3
	ModBanned   = "banned"
)

const (
	YearOfStudyOneID = 1
	YearOfStudyOne   = "one"

	YearOfStudyTwoID = 2
	YearOfStudyTwo   = "two"

	YearOfStudyThreeID = 3
	YearOfStudyThree   = "three"

	YearOfStudyFourID = 4
	YearOfStudyFour   = "four"

	YearOfStudyAlumniGraduateID = 5
	YearOfStudyAlumniGraduate   = "alumni_graduate"

	YearOfStudyHiddenID = 6
	YearOfStudyHidden   = "hidden"
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
}

type Faculty struct {
	ID      int    `gorm:"primaryKey" json:"id"`
	Faculty string `gorm:"column:faculty" json:"faculty"`
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

type User struct {
	ID          string     `gorm:"primaryKey" json:"id"`
	CreatedAt   TimeMicros `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   TimeMicros `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	YearOfStudy uint8      `gorm:"column:year_of_study" json:"year_of_study"`
	FacultyID   uint       `gorm:"column:faculty_id" json:"faculty_id"`
	SchoolID    uint       `gorm:"column:school_id" json:"school_id"`
	ModID       uint       `gorm:"column:mod_id" json:"mod_id"`
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
	FacultyID     uint            `gorm:"column:faculty_id" json:"-"`
	Faculty       Faculty         `gorm:"foreignKey:FacultyID" json:"faculty"`
	Title         string          `gorm:"column:title" json:"title"`
	Content       string          `gorm:"column:content" json:"content"`
	Downvote      uint            `gorm:"column:downvote" json:"downvote"`
	Upvote        uint            `gorm:"column:upvote" json:"upvote"`
	VoteScore     int             `gorm:"column:vote_score" json:"-"` // redundant to return to the user
	TrendingScore float64         `gorm:"column:trending_score" json:"trending_score"`
	HottestOn     *datatypes.Date `gorm:"column:hottest_on" json:"hottest_on"` // intentionally a pointer, so that it defaults to NULL when created and not specified (i.e. not its zero-value)
	Hidden        bool            `gorm:"column:hidden" json:"-"`
}

// ! Very important that SOME FIELDS ARE NOT EVER SERIALIZED TO PROTECT SENSATIVE DATA (json:"-")
type Comment struct {
	ID            int                `gorm:"primary_key;column:id" json:"id"`
	CreatedAt     TimeMicros         `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     TimeMicros         `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	PostID        uint               `gorm:"column:post_id" json:"post_id"`
	IdentifierID  uint               `gorm:"column:identifier_id" json:"-"`
	Identifier    *CommentIdentifier `gorm:"foreignKey:IdentifierID" json:"identifier"`
	Ancestors     pq.Int64Array      `gorm:"type:integer[]" json:"ancestors"`
	ChildrenCount uint               `gorm:"column:children_count" json:"children_count"`
	UserID        string             `gorm:"column:user_id" json:"-"`
	Content       string             `gorm:"column:content" json:"content"`
	Downvote      uint               `gorm:"column:downvote" json:"downvote"`
	Upvote        uint               `gorm:"column:upvote" json:"upvote"`
	VoteScore     int                `gorm:"column:vote_score" json:"-"` // redundant to return to the user
	TrendingScore float64            `gorm:"column:trending_score" json:"trending_score"`
	Hidden        bool               `gorm:"column:hidden" json:"-"`
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
	PostID    uint   `db:"post_id" gorm:"default:NULL" json:"post_id"`       // Either one of these FKs can be null, but the constraint
	CommentID uint   `db:"comment_id" gorm:"default:NULL" json:"comment_id"` // is that exactly one of them is a valid FK
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
	ID        int        `gorm:"primary_key;column:id" json:"id"`
	CreatedAt TimeMicros `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UserID    string     `gorm:"column:user_id" json:"-"`
	Content   string     `gorm:"column:content" json:"content"`
	TypeID    uint       `gorm:"column:type_id" json:"type_id"` // references the feedback_type table
}

type FeedbackType struct {
	ID   int    `gorm:"primary_key;column:id" json:"id"`
	Type string `gorm:"column:type" json:"type"`
}

func (FeedbackType) TableName() string {
	return "feedback_types"
}

// ! Important not to serialize some fields!!
type Report struct {
	UserID      string `gorm:"column:user_id" json:"-"`
	Description string `gorm:"column:description" json:"description"`
	Type        string `gorm:"type" json:"type"`
	Result      string `gorm:"column:result" json:"result"`
	UserAlerted bool   `gorm:"column:user_alerted" json:"user_alerted"`
}

type DailyHottestCron struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	SuccessfullyRan datatypes.Date `gorm:"column:successfully_ran" json:"successfully_ran"`
}

func (DailyHottestCron) TableName() string {
	return "daily_hottest_cron_jobs"
}

// ! Very important that SOME FIELDS ARE NOT EVER SERIALIZED TO PROTECT SENSATIVE DATA (json:"-")
// only serialize the fields that are needed for the client (if OP or identifier)
type CommentIdentifier struct {
	ID         uint       `gorm:"primaryKey" json:"-"`
	CreatedAt  TimeMicros `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  TimeMicros `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	UserID     string     `gorm:"column:user_id" json:"-"`
	PostID     uint       `gorm:"column:post_id" json:"-"`
	IsOp       bool       `gorm:"column:is_op" json:"is_op"`
	Identifier *int64     `gorm:"column:identifier" json:"identifier"` // pointer so it can be nullable
}

func (CommentIdentifier) TableName() string {
	return "comment_identifiers"
}
