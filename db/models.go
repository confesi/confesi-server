package db

import (
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Table names
const (
	Comments      = "comments"
	Faculties     = "faculties"
	Feedbacks     = "feedbacks"
	ModLevels     = "mod_levels"
	Posts         = "posts"
	Reports       = "reports"
	SavedComments = "saved_comments"
	SavedPosts    = "saved_posts"
	SchoolFollows = "school_follows"
	Schools       = "schools"
	Users         = "users"
	Votes         = "votes"
)

const (
	ModEnable   = "enabled"
	ModEnableID = 1

	ModBannedID = 2
	ModBanned   = "banned"

	ModLimitedID = 3
	ModLimited   = "limited"
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
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ModLevel struct {
	ID  uint
	Mod string
}

type School struct {
	ID     uint
	Name   string
	Abbr   string
	Lat    float32
	Lon    float32
	Domain string
}

type Faculty struct {
	ID      int
	Faculty string
}

func (Faculty) TableName() string {
	return "faculties"
}

type User struct {
	ID          string `gorm:"primaryKey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Email       string
	YearOfStudy uint8
	FacultyID   uint
	SchoolID    uint
	ModID       uint
}

type SchoolFollow struct {
	ID       uint
	UserID   uint
	SchoolID uint
}

type Post struct {
	CreatedAt     time.Time    `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time    `gorm:"column:updated_at;autoUpdateTime"`
	UserID        string       `gorm:"column:user_id"`
	SchoolID      uint         `gorm:"column:school_id"`
	FacultyID     uint         `gorm:"column:faculty_id"`
	Title         string       `gorm:"column:title"`
	Content       string       `gorm:"column:content"`
	Downvote      uint         `gorm:"column:downvote"`
	Upvote        uint         `gorm:"column:upvote"`
	TrendingScore uint64       `gorm:"column:trending_score"`
	HottestOn     sql.NullTime `gorm:"column:hottest_on"`
	Hidden        bool         `gorm:"column:hidden"`
	VoteScore     int          `gorm:"column:vote_score"`
}

// Implementing an interface
// GORM hook that will be called after updating a post (auto-update the `vote_score` + `trending_score` fields)
func (p *Post) AfterUpdate(tx *gorm.DB) error {
	fmt.Println("=======> GORM HOOK CALLED!!")
	p.VoteScore = int(p.Upvote) - int(p.Downvote)
	// todo: add trending score calculation
	return nil
}

// todo: make it implement both the post and comment interface

type Comment struct {
	meta
	UserID    string
	PostID    string
	CommentID *uint
	Content   string
	Downvote  uint
	Upvote    uint
	Score     int
	Hidden    bool
}

const (
	Upvote   = 1
	Downvote = -1
)

type Vote struct {
	ID        uint
	Vote      int    `db:"vote"`
	UserID    string `db:"user_id"`
	PostID    uint   `db:"post_id" gorm:"default:NULL"`    // Either one of these FKs can be null, but the constraint
	CommentID uint   `db:"comment_id" gorm:"default:NULL"` // is that exactly one of them is a valid FK
}

type SavedPost struct {
	meta
	UserID string
	PostID string
}

type SavedComment struct {
	meta
	UserID    string
	CommentID uint
}

type Feedback struct {
	meta
	UserID  string
	Content string
}

type Report struct {
	meta
	UserID      string
	Description string
	ReportType  string `gorm:"type"`
	Result      string
	UserAlerted bool
}
