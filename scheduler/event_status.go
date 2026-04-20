package scheduler

import (
	"log"
	"time"

	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

// StartEventStatusScheduler keeps event status in sync with time.
// Every 15 minutes it marks events as completed when start_time has passed,
// excluding events that are already completed or cancelled.
func StartEventStatusScheduler(db *gorm.DB) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	log.Println("🕒 Event status scheduler started (runs every 15 minutes)")

	// Run once immediately on startup to catch stale records.
	RunEventStatusSyncOnce(db)

	for range ticker.C {
		RunEventStatusSyncOnce(db)
	}
}

func RunEventStatusSyncOnce(db *gorm.DB) {
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
		return
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
		return
	}

	if completedResult.RowsAffected > 0 {
		log.Printf("🕒 Auto-completed %d event(s)", completedResult.RowsAffected)
	}
}
