package scheduler

import (
	"log"
	"time"

	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/service"
	"gorm.io/gorm"
)

// StartEventReminderScheduler runs a ticker every 30 minutes.
// It finds EventRegistrations for events starting between now+23h and now+25h
// where ReminderSent = false, sends a notification, and marks ReminderSent = true.
//
// Call via: go scheduler.StartEventReminderScheduler(db, notifSvc)
func StartEventReminderScheduler(db *gorm.DB, notifSvc service.NotificationService) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	log.Println("🔔 Event reminder scheduler started (runs every 30 minutes)")

	// Run once immediately on startup to catch any missed reminders
	RunEventRemindersOnce(db, notifSvc)

	for range ticker.C {
		RunEventRemindersOnce(db, notifSvc)
	}
}

func RunEventRemindersOnce(db *gorm.DB, notifSvc service.NotificationService) {
	now := time.Now()
	windowStart := now.Add(23 * time.Hour)
	windowEnd := now.Add(25 * time.Hour)

	// Find registrations for events starting in the 23-25h window that haven't been reminded
	type reminderRow struct {
		RegistrationID uint
		UserID         uint
		EventID        uint
		EventTitle     string
	}

	var rows []reminderRow
	err := db.Raw(`
		SELECT er.id AS registration_id, er.user_id, e.id AS event_id, e.title AS event_title
		FROM event_registrations er
		JOIN events e ON e.id = er.event_id AND e.deleted_at IS NULL
		WHERE er.reminder_sent = false
		  AND er.deleted_at IS NULL
		  AND e.start_time >= ?
		  AND e.start_time < ?
	`, windowStart, windowEnd).Scan(&rows).Error

	if err != nil {
		log.Printf("⚠️  Event reminder scheduler error querying DB: %v", err)
		return
	}

	if len(rows) == 0 {
		return
	}

	log.Printf("🔔 Sending %d event reminder(s)...", len(rows))

	for _, row := range rows {
		body := row.EventTitle + " dimulai dalam 24 jam"
		if err := notifSvc.Notify(
			row.UserID,
			"event_reminder",
			"Pengingat event",
			body,
			"event",
			row.EventID,
		); err != nil {
			log.Printf("⚠️  Failed to send reminder for reg %d: %v", row.RegistrationID, err)
			continue
		}

		// Mark as sent to prevent duplicate reminders
		if err := db.Model(&models.EventRegistration{}).
			Where("id = ?", row.RegistrationID).
			Update("reminder_sent", true).Error; err != nil {
			log.Printf("⚠️  Failed to mark reminder_sent for reg %d: %v", row.RegistrationID, err)
		}
	}
}
