package service

import (
	"log"
	"time"

	"github.com/reganputra/skripsi-backend/repository"
)

// ─── DTOs ─────────────────────────────────────────────────────────────────────

// OverviewResponse is returned by GET /api/admin/analytics/overview.
type OverviewResponse struct {
	TotalReactions int64             `json:"total_reactions"`
	TotalComments  int64             `json:"total_comments"`
	NewUsersLast7d int64             `json:"new_users_last_7d"`
	Yesterday      *OverviewSnapshot `json:"yesterday,omitempty"`
}

// OverviewSnapshot carries the latest snapshot cards shown on the dashboard.
type OverviewSnapshot struct {
	ActiveUsers int64 `json:"active_users"`
	PostViews   int64 `json:"post_views"`
	GroupViews  int64 `json:"group_views"`
}

// TrendPoint is a narrow time-series row for the admin chart.
type TrendPoint struct {
	Date        string `json:"date"`
	ActiveUsers int64  `json:"active_users"`
	NewUsers    int64  `json:"new_users"`
	PostViews   int64  `json:"post_views"`
	GroupViews  int64  `json:"group_views"`
}

// TopContentItem enriches a TopContentRow with human-readable content info.
type TopContentItem struct {
	TargetID  uint  `json:"target_id"`
	ViewCount int64 `json:"view_count"`
}

// ─── Interface ────────────────────────────────────────────────────────────────

type AnalyticsService interface {
	// GetOverview returns lifetime stats + yesterday's snapshot.
	GetOverview() (*OverviewResponse, error)
	// GetTrends returns up to `days` daily snapshots ordered by date asc.
	GetTrends(days int) ([]TrendPoint, error)
	// GetTopContent returns the top N items of a given type by raw view count.
	GetTopContent(targetType string, days, limit int) ([]TopContentItem, error)
	// BuildSnapshotForDate computes and persists a snapshot for a single day.
	BuildSnapshotForDate(date time.Time) error
	// BackfillSnapshots computes and upserts snapshots for the last `days` days.
	BackfillSnapshots(days int) error
}

// ─── Implementation ───────────────────────────────────────────────────────────

type analyticsService struct {
	repo repository.AnalyticsRepository
}

func NewAnalyticsService(repo repository.AnalyticsRepository) AnalyticsService {
	return &analyticsService{repo: repo}
}

func (s *analyticsService) GetOverview() (*OverviewResponse, error) {
	stats, err := s.repo.GetEnhancedStats()
	if err != nil {
		return nil, err
	}

	snaps, err := s.repo.GetDailySnapshots(1)
	var yesterdaySnap *OverviewSnapshot
	if err == nil && len(snaps) > 0 {
		y := snaps[len(snaps)-1]
		yesterdaySnap = &OverviewSnapshot{
			ActiveUsers: y.ActiveUsers,
			PostViews:   y.PostViews,
			GroupViews:  y.GroupViews,
		}
	}

	return &OverviewResponse{
		TotalReactions: stats.TotalReactions,
		TotalComments:  stats.TotalComments,
		NewUsersLast7d: stats.NewUsersLast7d,
		Yesterday:      yesterdaySnap,
	}, nil
}

func (s *analyticsService) GetTrends(days int) ([]TrendPoint, error) {
	snaps, err := s.repo.GetDailySnapshots(days)
	if err != nil {
		return nil, err
	}
	points := make([]TrendPoint, 0, len(snaps))
	for _, snap := range snaps {
		points = append(points, TrendPoint{
			Date:        snap.Date.Format("2006-01-02"),
			ActiveUsers: snap.ActiveUsers,
			NewUsers:    snap.NewUsers,
			PostViews:   snap.PostViews,
			GroupViews:  snap.GroupViews,
		})
	}
	return points, nil
}

func (s *analyticsService) GetTopContent(targetType string, days, limit int) ([]TopContentItem, error) {
	rows, err := s.repo.GetTopContent(targetType, days, limit)
	if err != nil {
		return nil, err
	}
	items := make([]TopContentItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, TopContentItem{
			TargetID:  r.TargetID,
			ViewCount: r.ViewCount,
		})
	}
	return items, nil
}

func (s *analyticsService) BuildSnapshotForDate(date time.Time) error {
	snap, err := s.repo.ComputeSnapshotForDate(date)
	if err != nil {
		return err
	}
	return s.repo.UpsertSnapshot(snap)
}

// BackfillSnapshots computes and upserts daily snapshots for the last `days`
// calendar days. This is run once at startup if snapshots are missing, so trend
// charts have historical data from day one of deployment.
func (s *analyticsService) BackfillSnapshots(days int) error {
	if days <= 0 {
		days = 90
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	log.Printf("[analytics] backfill: computing %d days of snapshots...", days)
	for i := days; i >= 1; i-- {
		date := today.AddDate(0, 0, -i)
		if err := s.BuildSnapshotForDate(date); err != nil {
			log.Printf("[analytics] backfill: error for %s: %v", date.Format("2006-01-02"), err)
			continue // don't abort; best-effort
		}
	}
	log.Printf("[analytics] backfill: complete")
	return nil
}
