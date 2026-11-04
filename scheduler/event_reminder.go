package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/service"
	"gorm.io/gorm"
)

const (
	reminderBatchSize          = 100
	schedulerRunTimeout        = 30 * time.Second
	reminderTransactionTimeout = 5 * time.Second
)

type reminderRow struct {
	RegistrationID uint
	UserID         uint
	EventID        uint
	EventTitle     string
}

// StartEventReminderScheduler cancels in-flight DB work as well as future ticks.
func StartEventReminderScheduler(ctx context.Context, db *gorm.DB, notifSvc service.NotificationService) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil {
			return
		}
		if err := RunEventRemindersOnceContext(ctx, db, notifSvc); err != nil {
			log.Printf("[event-reminder] %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// RunEventRemindersOnce retains the original API; new callers should pass a context.
func RunEventRemindersOnce(db *gorm.DB, notifSvc service.NotificationService) {
	if err := RunEventRemindersOnceContext(context.Background(), db, notifSvc); err != nil {
		log.Printf("[event-reminder] %v", err)
	}
}

// RunEventRemindersOnceContext persists reminders and their sent flags atomically.
// A publish-only capability delivers after commit without inserting duplicates.
// A crash after commit may miss the realtime hint; the REST copy remains durable.
func RunEventRemindersOnceContext(ctx context.Context, db *gorm.DB, notifSvc service.NotificationService) error {
	ctx, cancel := context.WithTimeout(ctx, schedulerRunTimeout)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return err
	}
	now := time.Now()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		count, err := persistReminderBatch(ctx, db, now, func(n *models.Notification) {
			if publisher, ok := notifSvc.(interface{ PublishPersisted(*models.Notification) }); ok {
				publisher.PublishPersisted(n)
			}
		})
		if err != nil {
			return err
		}
		if count < reminderBatchSize {
			return nil
		}
	}
}

func reminderNotification(row reminderRow) models.Notification {
	return models.Notification{
		UserID: row.UserID, NotificationType: "event_reminder", Title: "Pengingat event",
		Body: row.EventTitle + " dimulai dalam 24 jam", ReferenceType: "event",
		ReferenceID: row.EventID, RedirectURL: fmt.Sprintf("/events/%d", row.EventID),
	}
}

// Short transactions lock only registrations so replicas process disjoint batches.
// No session locks are used, making this safe with transaction poolers.
func persistReminderBatch(ctx context.Context, db *gorm.DB, now time.Time, publish ...func(*models.Notification)) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, reminderTransactionTimeout)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	var rows []reminderRow
	var notifications []models.Notification
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(`
   SELECT er.id AS registration_id, er.user_id, e.id AS event_id, e.title AS event_title
   FROM event_registrations er
   JOIN events e ON e.id = er.event_id AND e.deleted_at IS NULL
   WHERE er.reminder_sent = false AND er.deleted_at IS NULL
     AND e.start_time >= ? AND e.start_time < ?
   ORDER BY er.id
   LIMIT ? FOR UPDATE OF er SKIP LOCKED`,
			now.Add(23*time.Hour), now.Add(25*time.Hour), reminderBatchSize).Scan(&rows).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		notifications = make([]models.Notification, 0, len(rows))
		ids := make([]uint, 0, len(rows))
		for _, row := range rows {
			notifications = append(notifications, reminderNotification(row))
			ids = append(ids, row.RegistrationID)
		}
		if err := tx.Create(&notifications).Error; err != nil {
			return err
		}
		result := tx.Model(&models.EventRegistration{}).
			Where("id IN ? AND reminder_sent = false", ids).Update("reminder_sent", true)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != int64(len(rows)) {
			return fmt.Errorf("reminder flags updated %d rows, expected %d", result.RowsAffected, len(rows))
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	for i := range notifications {
		if ctx.Err() != nil {
			break
		}
		for _, deliver := range publish {
			if deliver != nil {
				deliver(&notifications[i])
			}
		}
	}
	return len(rows), nil
}
