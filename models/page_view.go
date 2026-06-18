package models

import "time"

// PageView is an append-only event log recording each authenticated user's
// visit to a content detail page (post, event, job, or group).
//
// Design notes:
//   - No gorm.Model — views are immutable facts; soft-delete and UpdatedAt are unnecessary.
//   - UserID is nullable: currently all routes require auth, so it will always
//     be populated, but nil is allowed to support future public pages gracefully.
//   - TargetType mirrors the pattern used by Report ("post"|"event"|"job"|"group").
//   - Indexed on (target_type, target_id) for per-item counts and on
//     created_at for time-range aggregations.
//   - idx_pv_cooldown composite index on (user_id, target_type, target_id, created_at)
//     is used exclusively by the deduplication gate in the TrackView middleware
//     to determine whether a user has already viewed the same content within the
//     cooldown window, without a sequential scan.
type PageView struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	CreatedAt  time.Time `gorm:"index;not null"`
	UserID     *uint     `gorm:"index;index:idx_pv_cooldown,priority:1"`
	TargetType string    `gorm:"not null;index:idx_pv_target;index:idx_pv_cooldown,priority:2"`
	TargetID   uint      `gorm:"not null;index:idx_pv_target;index:idx_pv_cooldown,priority:3"`
}
