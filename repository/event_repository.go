package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type EventRepository interface {
	CreateEvent(event *models.Event) error
	FindEvents(offset, limit int) ([]models.Event, int64, error)
	FindEventByID(id uint) (*models.Event, error)
	UpdateEvent(event *models.Event) error
	DeleteEvent(id uint) error
	CountEventRegistrations(eventID uint) (int64, error)
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

func (r *eventRepository) FindEventByID(id uint) (*models.Event, error) {
	var event models.Event
	err := r.db.
		Preload("User").
		Preload("Agendas").
		Preload("Registrations.User").
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
