package repository

import (
	"time"

	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type EventRepository interface {
	CreateEvent(event *models.Event) error
	FindEvents(offset, limit int) ([]models.Event, int64, error)
	FindEventsByOwner(userID uint, offset, limit int) ([]models.Event, int64, error)
	FindEventByID(id uint) (*models.Event, error)
	UpdateEvent(event *models.Event) error
	DeleteEvent(id uint) error
	CountEventRegistrations(eventID uint) (int64, error)
	// CountEventRegistrationsBatch returns a map[eventID]count for a slice of
	// event IDs fetched in a single GROUP BY query (avoids N+1).
	CountEventRegistrationsBatch(eventIDs []uint) (map[uint]int64, error)
	AutoCompletePastEvents(now time.Time) (int64, error)
}

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) CreateEvent(event *models.Event) error {
	return r.db.Create(event).Error
}

func (r *eventRepository) FindEvents(offset, limit int) ([]models.Event, int64, error) {
	var events []models.Event
	var total int64
	r.db.Model(&models.Event{}).Count(&total)
	err := r.db.
		Preload("User").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&events).Error
	return events, total, err
}

func (r *eventRepository) FindEventsByOwner(userID uint, offset, limit int) ([]models.Event, int64, error) {
	var events []models.Event
	var total int64

	base := r.db.Model(&models.Event{}).Where("user_id = ?", userID)
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.
		Preload("User").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&events).Error

	return events, total, err
}

func (r *eventRepository) FindEventByID(id uint) (*models.Event, error) {
	var event models.Event
	err := r.db.
		Preload("User").
		Preload("User.Profile").
		Preload("Agendas").
		Preload("Registrations.User").
		Preload("Registrations.User.Profile").
		First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) UpdateEvent(event *models.Event) error {
	return r.db.Save(event).Error
}

func (r *eventRepository) DeleteEvent(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("event_id = ?", id).Delete(&models.EventAgenda{}).Error; err != nil {
			return err
		}
		if err := tx.Where("event_id = ?", id).Delete(&models.EventRegistration{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Event{}, id).Error
	})
}

func (r *eventRepository) CountEventRegistrations(eventID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.EventRegistration{}).Where("event_id = ?", eventID).Count(&count).Error
	return count, err
}

// CountEventRegistrationsBatch returns registration counts for multiple events
// in a single query, avoiding N+1 in list endpoints.
func (r *eventRepository) CountEventRegistrationsBatch(eventIDs []uint) (map[uint]int64, error) {
	if len(eventIDs) == 0 {
		return map[uint]int64{}, nil
	}

	type countRow struct {
		EventID uint  `gorm:"column:event_id"`
		Cnt     int64 `gorm:"column:cnt"`
	}

	var rows []countRow
	err := r.db.Raw(`
		SELECT event_id, COUNT(*) AS cnt
		FROM event_registrations
		WHERE event_id IN (?) AND deleted_at IS NULL
		GROUP BY event_id
	`, eventIDs).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint]int64, len(rows))
	for _, row := range rows {
		result[row.EventID] = row.Cnt
	}
	return result, nil
}

func (r *eventRepository) AutoCompletePastEvents(now time.Time) (int64, error) {
	fallbackCompletionTime := now.Add(-24 * time.Hour)

	// Transition: upcoming → ongoing when now >= start_time
	ongoingResult := r.db.Model(&models.Event{}).
		Where("start_time IS NOT NULL").
		Where("start_time <= ?", now).
		Where("status = ?", "upcoming").
		Update("status", "ongoing")

	if ongoingResult.Error != nil {
		return 0, ongoingResult.Error
	}

	// Transition: ongoing → completed when now >= end_time (or start_time + 24h if no end_time)
	completedResult := r.db.Model(&models.Event{}).
		Where("start_time IS NOT NULL").
		Where("status = ?", "ongoing").
		Where(r.db.Where("end_time IS NOT NULL AND end_time <= ?", now).
			Or("end_time IS NULL AND start_time <= ?", fallbackCompletionTime)).
		Update("status", "completed")

	if completedResult.Error != nil {
		return ongoingResult.RowsAffected, completedResult.Error
	}

	return ongoingResult.RowsAffected + completedResult.RowsAffected, nil
}
