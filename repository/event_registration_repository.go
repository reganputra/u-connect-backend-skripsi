package repository

import (
	"errors"

	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type EventRegistrationRepository interface {
	CreateEventRegistration(reg *models.EventRegistration) error
	FindEventRegistration(eventID, userID uint) (*models.EventRegistration, error)
	DeleteEventRegistration(eventID, userID uint) error
	FindEventParticipants(eventID uint) ([]models.EventRegistration, error)
	FindRegisteredEventsByUser(userID uint, offset, limit int) ([]models.Event, int64, error)
}

type eventRegistrationRepository struct {
	db *gorm.DB
}

func NewEventRegistrationRepository(db *gorm.DB) EventRegistrationRepository {
	return &eventRegistrationRepository{db: db}
}

func (r *eventRegistrationRepository) CreateEventRegistration(reg *models.EventRegistration) error {
	var existing models.EventRegistration
	err := r.db.Unscoped().
		Where("event_id = ? AND user_id = ?", reg.EventID, reg.UserID).
		First(&existing).Error

	if err == nil {
		if existing.DeletedAt.Valid {
			return r.db.Unscoped().
				Model(&models.EventRegistration{}).
				Where("id = ?", existing.ID).
				Updates(map[string]interface{}{
					"deleted_at":    nil,
					"reminder_sent": false,
				}).Error
		}

		// Let DB enforce uniqueness for concurrent duplicate registrations.
		return r.db.Create(reg).Error
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return r.db.Create(reg).Error
}

func (r *eventRegistrationRepository) FindEventRegistration(eventID, userID uint) (*models.EventRegistration, error) {
	var reg models.EventRegistration
	err := r.db.Where("event_id = ? AND user_id = ?", eventID, userID).First(&reg).Error
	if err != nil {
		return nil, err
	}
	return &reg, nil
}

func (r *eventRegistrationRepository) DeleteEventRegistration(eventID, userID uint) error {
	return r.db.Unscoped().
		Where("event_id = ? AND user_id = ?", eventID, userID).
		Delete(&models.EventRegistration{}).Error
}

func (r *eventRegistrationRepository) FindEventParticipants(eventID uint) ([]models.EventRegistration, error) {
	var regs []models.EventRegistration
	err := r.db.
		Preload("User").
		Where("event_id = ?", eventID).
		Find(&regs).Error
	return regs, err
}

func (r *eventRegistrationRepository) FindRegisteredEventsByUser(userID uint, offset, limit int) ([]models.Event, int64, error) {
	var events []models.Event
	var total int64

	countQuery := r.db.Model(&models.Event{}).
		Joins("JOIN event_registrations ON event_registrations.event_id = events.id").
		Where("event_registrations.user_id = ?", userID)

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Model(&models.Event{}).
		Joins("JOIN event_registrations ON event_registrations.event_id = events.id").
		Where("event_registrations.user_id = ?", userID).
		Preload("User").
		Order("event_registrations.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&events).Error

	return events, total, err
}
