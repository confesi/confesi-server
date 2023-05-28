package models

import (
	"time"
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
	Email       string `json:"email"`
	YearOfStudy uint8  `json:"yearOfStudy"`
	FacultyID   uint   `json:"-"`
	SchoolID    uint   `json:"-"`
	ModID       uint   `json:"-"`
}

type SchoolFollow struct {
	ID       uint
	UserID   uint
	SchoolID uint
}

type Post struct {
	meta
	UserID        string
	SchoolID      uint
	FacultyID     uint
	Title         string `json:"title"`
	Content       string `json:"content"`
	Downvote      uint   `json:"downvote"`
	Upvote        uint   `json:"upvote"`
	TrendingScore uint64 `json:"trendingScore"`
	HottestScore  uint64 `json:"hottestScore"`
	Hidden        bool   `json:"hidden"`
}

type Comment struct {
	meta
	UserID    string
	PostID    string
	CommentID *uint
	Content   string
	Downvote  uint `json:"downvote"`
	Upvote    uint `json:"upvote"`
	Score     int  `json:"score"`
	Hidden    bool `json:"hidden"`
}

const (
	Upvote   = 1
	Downvote = -1
)

type Vote struct {
	ID        uint
	Vote      int
	UserID    string
	PostID    uint
	CommentID uint
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
