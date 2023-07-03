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

type meta struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt TimeMillis
	UpdatedAt TimeMillis
}

type ModLevel struct {
	ID  uint
	Mod string
}

type School struct {
	ID            uint    `json:"id"`
	Name          string  `json:"name"`
	Abbr          string  `json:"abbr"`
	Lat           float32 `json:"lat"`
	Lon           float32 `json:"lon"`
	DailyHottests int     `json:"daily_hottests"`
	Domain        string  `json:"domain"`
}

type Faculty struct {
	ID      int
	Faculty string
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
	ID          string     `gorm:"primaryKey"`
	CreatedAt   TimeMillis `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   TimeMillis `gorm:"column:updated_at;autoUpdateTime"`
	YearOfStudy uint8
	FacultyID   uint
	SchoolID    uint
	ModID       uint
}

// ! Very important some fields are NOT serialized (json:"-")
type SchoolFollow struct {
	ID        uint       `gorm:"primary_key;column:id" json:"-"`
	CreatedAt TimeMillis `gorm:"column:created_at;autoCreateTime" json:"-"`
	UpdatedAt TimeMillis `gorm:"column:updated_at;autoUpdateTime" json:"-"`
	UserID    string     `gorm:"column:user_id" json:"-"`
	SchoolID  uint
}

// ! Very important that SOME FIELDS ARE NOT EVER SERIALIZED TO PROTECT SENSATIVE DATA (json:"-")
type Post struct {
	ID            int             `gorm:"primary_key;column:id" json:"-"`
	CreatedAt     TimeMillis      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     TimeMillis      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	UserID        string          `gorm:"column:user_id" json:"-"`
	SchoolID      uint            `gorm:"column:school_id" json:"-"`
	School        School          `gorm:"foreignKey:SchoolID"`
	FacultyID     uint            `gorm:"column:faculty_id" json:"-"`
	Faculty       Faculty         `gorm:"foreignKey:FacultyID"`
	Title         string          `gorm:"column:title"`
	Content       string          `gorm:"column:content"`
	Downvote      uint            `gorm:"column:downvote"`
	Upvote        uint            `gorm:"column:upvote"`
	TrendingScore float64         `gorm:"column:trending_score"`
	HottestOn     *datatypes.Date `gorm:"column:hottest_on"` // intentionally a pointer, so that it defaults to NULL when created and not specified (i.e. not its zero-value)
	Hidden        bool            `gorm:"column:hidden" json:"-"`
	VoteScore     int             `gorm:"column:vote_score"`
}

// ! Very important that SOME FIELDS ARE NOT EVER SERIALIZED TO PROTECT SENSATIVE DATA (json:"-")
type Comment struct {
	ID            int        `gorm:"primary_key;column:id"`
	CreatedAt     TimeMillis `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     TimeMillis `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	PostID        uint
	IdentifierID  uint               `gorm:"column:identifier_id" json:"-"`
	Identifier    *CommentIdentifier `gorm:"foreignKey:IdentifierID"`
	Ancestors     pq.Int64Array      `gorm:"type:integer[]"`
	ChildrenCount uint
	UserID        string `gorm:"column:user_id" json:"-"`
	Content       string
	Downvote      uint
	Upvote        uint
	Score         int
	Hidden        bool `gorm:"column:hidden" json:"-"`
}

// This will store as a `time.Time` in the database,
// and as a unix ms time for json serialization
type TimeMillis struct {
	time.Time
}

func (t TimeMillis) MarshalJSON() ([]byte, error) {
	millis := t.UnixMilli()
	return json.Marshal(millis)
}

func (t *TimeMillis) UnmarshalJSON(data []byte) error {
	var millis int64
	err := json.Unmarshal(data, &millis)
	if err != nil {
		return err
	}
	t.Time = time.Unix(0, millis*int64(time.Millisecond))
	return nil
}

func (t TimeMillis) Value() (driver.Value, error) {
	return t.Time, nil
}

func (t *TimeMillis) Scan(value interface{}) error {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		t.Time = v
	case []byte:
		var millis int64
		err := json.Unmarshal(v, &millis)
		if err != nil {
			return err
		}
		t.Time = time.Unix(0, millis*int64(time.Millisecond))
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

type Vote struct {
	ID        uint
	Vote      int    `db:"vote"`
	UserID    string `db:"user_id"`
	PostID    uint   `db:"post_id" gorm:"default:NULL"`    // Either one of these FKs can be null, but the constraint
	CommentID uint   `db:"comment_id" gorm:"default:NULL"` // is that exactly one of them is a valid FK
}

type SavedPost struct {
	CreatedAt TimeMillis `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt TimeMillis `gorm:"column:updated_at;autoUpdateTime"`
	UserID    string
	PostID    uint
}

type SavedComment struct {
	CreatedAt TimeMillis `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt TimeMillis `gorm:"column:updated_at;autoUpdateTime"`
	UserID    string
	CommentID uint
}

type Feedback struct {
	ID        int        `gorm:"primary_key;column:id"`
	CreatedAt TimeMillis `gorm:"column:created_at;autoCreateTime"`
	UserID    string
	Content   string
	TypeID    uint // references the feedback_type table
}

type FeedbackType struct {
	ID   int `gorm:"primary_key;column:id"`
	Type string
}

func (FeedbackType) TableName() string {
	return "feedback_types"
}

type Report struct {
	meta
	UserID      string
	Description string
	ReportType  string `gorm:"type"`
	Result      string
	UserAlerted bool
}

type DailyHottestCron struct {
	ID              uint           `gorm:"primaryKey"`
	SuccessfullyRan datatypes.Date `gorm:"column:successfully_ran"`
}

func (DailyHottestCron) TableName() string {
	return "daily_hottest_cron_jobs"
}

// ! Very important that SOME FIELDS ARE NOT EVER SERIALIZED TO PROTECT SENSATIVE DATA (json:"-")
// only serialize the fields that are needed for the client (if OP or identifier)
type CommentIdentifier struct {
	ID         uint       `gorm:"primaryKey" json:"-"`
	CreatedAt  TimeMillis `gorm:"column:created_at;autoCreateTime" json:"-"`
	UpdatedAt  TimeMillis `gorm:"column:updated_at;autoUpdateTime" json:"-"`
	UserID     string     `gorm:"column:user_id" json:"-"`
	PostID     uint       `gorm:"column:post_id" json:"-"`
	IsOp       bool       `gorm:"column:is_op"`
	Identifier *int64     `gorm:"column:identifier"` // pointer so it can be nullable
}

func (CommentIdentifier) TableName() string {
	return "comment_identifiers"
}
