package repository

import (
	"time"

	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ─── DTOs ─────────────────────────────────────────────────────────────────────

// TopContentRow holds a single item in the "top content by views" result.
type TopContentRow struct {
	TargetID  uint  `gorm:"column:target_id"`
	ViewCount int64 `gorm:"column:view_count"`
}

// EnhancedStats carries the concise analytics summary used by the admin dashboard.
type EnhancedStats struct {
	TotalReactions int64 `json:"total_reactions"`
	TotalComments  int64 `json:"total_comments"`
	NewUsersLast7d int64 `json:"new_users_last_7d"`
}

// ─── Interface ────────────────────────────────────────────────────────────────

type AnalyticsRepository interface {
	// Tier 1
	GetEnhancedStats() (*EnhancedStats, error)

	// Tier 2 — trend data
	GetDailySnapshots(days int) ([]models.DailyAnalyticsSnapshot, error)
	GetTopContent(targetType string, days, limit int) ([]TopContentRow, error)

	// Snapshot management
	ComputeSnapshotForDate(date time.Time) (*models.DailyAnalyticsSnapshot, error)
	UpsertSnapshot(snap *models.DailyAnalyticsSnapshot) error
}

// ─── Implementation ───────────────────────────────────────────────────────────

type analyticsRepository struct{ db *gorm.DB }

func NewAnalyticsRepository(db *gorm.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

// ─── Tier 1: Enhanced Stats ───────────────────────────────────────────────────

func (r *analyticsRepository) GetEnhancedStats() (*EnhancedStats, error) {
	s := &EnhancedStats{}
	db := r.db

	// Helper for simple counts.
	count := func(model interface{}, where ...string) (int64, error) {
		var n int64
		q := db.Model(model)
		for _, w := range where {
			q = q.Where(w)
		}
		err := q.Count(&n).Error
		return n, err
	}

	var err error
	s.TotalReactions, err = count(&models.Reaction{})
	if err != nil {
		return nil, err
	}
	s.TotalComments, err = count(&models.Comment{})
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	s.NewUsersLast7d, err = count(&models.User{}, "created_at >= '"+now.AddDate(0, 0, -7).Format(time.RFC3339)+"'")
	if err != nil {
		return nil, err
	}

	return s, nil
}

// ─── Tier 2: Trend Data ────────────────────────────────────────────────────────

func (r *analyticsRepository) GetDailySnapshots(days int) ([]models.DailyAnalyticsSnapshot, error) {
	if days <= 0 || days > 365 {
		days = 30
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -days)
	var snaps []models.DailyAnalyticsSnapshot
	err := r.db.
		Where("date >= ?", cutoff).
		Order("date asc").
		Find(&snaps).Error
	return snaps, err
}

func (r *analyticsRepository) GetTopContent(targetType string, days, limit int) ([]TopContentRow, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	if targetType != "post" && targetType != "group" {
		targetType = "post"
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -days)
	var rows []TopContentRow
	err := r.db.Model(&models.PageView{}).
		Select("target_id, COUNT(*) as view_count").
		Where("target_type = ? AND created_at >= ?", targetType, cutoff).
		Group("target_id").
		Order("view_count desc").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// ─── Snapshot Computation ─────────────────────────────────────────────────────

// ComputeSnapshotForDate aggregates metrics for the given UTC calendar day.
// It reads from real tables (page_views, users, posts, reactions, …) using
// time-bounded COUNT queries scoped to [dayStart, dayEnd).
func (r *analyticsRepository) ComputeSnapshotForDate(date time.Time) (*models.DailyAnalyticsSnapshot, error) {
	// Normalise to UTC midnight
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.AddDate(0, 0, 1)

	snap := &models.DailyAnalyticsSnapshot{Date: dayStart}
	db := r.db

	countBetween := func(model interface{}, col string) int64 {
		var n int64
		db.Model(model).Where(col+" >= ? AND "+col+" < ?", dayStart, dayEnd).Count(&n)
		return n
	}

	// Active users: distinct user_ids in page_views for this day.
	db.Model(&models.PageView{}).
		Select("COUNT(DISTINCT user_id)").
		Where("created_at >= ? AND created_at < ? AND user_id IS NOT NULL", dayStart, dayEnd).
		Scan(&snap.ActiveUsers)

	snap.NewUsers = countBetween(&models.User{}, "created_at")

	// Page views split by target type.
	viewCount := func(targetType string) int64 {
		var n int64
		db.Model(&models.PageView{}).
			Where("target_type = ? AND created_at >= ? AND created_at < ?", targetType, dayStart, dayEnd).
			Count(&n)
		return n
	}
	snap.PostViews = viewCount("post")
	snap.GroupViews = viewCount("group")

	return snap, nil
}

// UpsertSnapshot inserts or updates a DailyAnalyticsSnapshot by its unique Date.
func (r *analyticsRepository) UpsertSnapshot(snap *models.DailyAnalyticsSnapshot) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"active_users", "new_users", "post_views", "group_views",
		}),
	}).Create(snap).Error
}
