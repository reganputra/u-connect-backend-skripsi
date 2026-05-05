package models

import "time"

// DailyAnalyticsSnapshot is a pre-aggregated daily rollup used for trend charts.
// The snapshot scheduler runs each night and inserts/updates the row for the
// previous calendar day. Dashboard queries read from this table instead of
// scanning large event tables on every request.
//
// Date is stored as a UTC midnight timestamp (time.Time truncated to day).
// The uniqueIndex on Date prevents duplicate rows per day.
type DailyAnalyticsSnapshot struct {
	ID   uint      `gorm:"primaryKey;autoIncrement"`
	Date time.Time `gorm:"uniqueIndex;not null"` // UTC midnight of the day

	// User activity
	ActiveUsers int64 `gorm:"not null;default:0"` // distinct user_ids in page_views for this day
	NewUsers    int64 `gorm:"not null;default:0"` // users.created_at falls on this day

	// Content created that day
	NewPosts    int64 `gorm:"not null;default:0"`
	NewEvents   int64 `gorm:"not null;default:0"`
	NewJobs     int64 `gorm:"not null;default:0"`
	NewArticles int64 `gorm:"not null;default:0"` // group articles

	// Engagement actions that day
	NewRegistrations int64 `gorm:"not null;default:0"` // event registrations
	NewApplications  int64 `gorm:"not null;default:0"` // job applications
	NewFollows       int64 `gorm:"not null;default:0"`
	NewMessages      int64 `gorm:"not null;default:0"`
	TotalReactions   int64 `gorm:"not null;default:0"`
	TotalComments    int64 `gorm:"not null;default:0"`
	TotalVotes       int64 `gorm:"not null;default:0"`

	// Page views by content type (raw count — deduplication is done at query time)
	PostViews  int64 `gorm:"not null;default:0"`
	EventViews int64 `gorm:"not null;default:0"`
	JobViews   int64 `gorm:"not null;default:0"`
	GroupViews int64 `gorm:"not null;default:0"`
}
