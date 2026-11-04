package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

// StartEventStatusScheduler keeps event status in sync with time.
// Every 15 minutes it starts due events and completes events at end_time
// (or start_time + 24h). Completed and cancelled events are left unchanged.
// The scheduler stops when ctx is cancelled (graceful shutdown).
func StartEventStatusScheduler(ctx context.Context, db *gorm.DB) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	log.Println("🕒 Event status scheduler started (runs every 15 minutes)")

	for {
		if ctx.Err() != nil {
			return
		}
		if err := RunEventStatusSyncOnceContext(ctx, db); err != nil {
			log.Printf("[event-status] %v", err)
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}

// RunEventStatusSyncOnce preserves the existing entry point.
func RunEventStatusSyncOnce(db *gorm.DB) {
	if err := RunEventStatusSyncOnceContext(context.Background(), db); err != nil {
		log.Printf("[event-status] %v", err)
	}
}

// RunEventStatusSyncOnceContext uses conditional, idempotent UPDATEs: PostgreSQL
// rechecks each status predicate after row-lock waits. A leader/session advisory
// lock is unnecessary because there are no external side effects. Separate
// statements converge on retry if cancellation interrupts the two transitions.
func RunEventStatusSyncOnceContext(ctx context.Context, db *gorm.DB) error {
	ctx, cancel := context.WithTimeout(ctx, schedulerRunTimeout)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return err
	}
	db = db.WithContext(ctx)
	now := time.Now()
	fallbackCompletionTime := now.Add(-24 * time.Hour)

	// Transition: upcoming → ongoing when now >= start_time
	ongoingResult := db.Model(&models.Event{}).
		Where("start_time IS NOT NULL").
		Where("start_time <= ?", now).
		Where("status = ?", "upcoming").
		Update("status", "ongoing")

	if ongoingResult.Error != nil {
		log.Printf("⚠️  Event status scheduler error (ongoing transition): %v", ongoingResult.Error)
		return ongoingResult.Error
	}

	if ongoingResult.RowsAffected > 0 {
		log.Printf("🕒 Auto-transitioned %d event(s) to ongoing", ongoingResult.RowsAffected)
	}

	// Transition: ongoing → completed when now >= end_time (or start_time + 24h if no end_time)
	completedResult := db.Model(&models.Event{}).
		Where("start_time IS NOT NULL").
		Where("status = ?", "ongoing").
		Where(db.Where("end_time IS NOT NULL AND end_time <= ?", now).
			Or("end_time IS NULL AND start_time <= ?", fallbackCompletionTime)).
		Update("status", "completed")

	if completedResult.Error != nil {
		log.Printf("⚠️  Event status scheduler error (completed transition): %v", completedResult.Error)
		return completedResult.Error
	}

	if completedResult.RowsAffected > 0 {
		log.Printf("🕒 Auto-completed %d event(s)", completedResult.RowsAffected)
	}
	return nil
}
