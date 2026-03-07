package repository

import (
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

func CreateEvent(event *models.Event) error {
	return config.DB.Create(event).Error
}

func FindEvents(offset, limit int) ([]models.Event, int64, error) {
	var events []models.Event
	var total int64
	config.DB.Model(&models.Event{}).Count(&total)
	err := config.DB.
		Preload("User").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&events).Error
	return events, total, err
}

func FindEventByID(id uint) (*models.Event, error) {
	var event models.Event
	err := config.DB.
		Preload("User").
		Preload("Agendas").
		Preload("Registrations.User").
		First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func UpdateEvent(event *models.Event) error {
	return config.DB.Save(event).Error
}

func DeleteEvent(id uint) error {
	return config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("event_id = ?", id).Delete(&models.EventAgenda{}).Error; err != nil {
			return err
		}
		if err := tx.Where("event_id = ?", id).Delete(&models.EventRegistration{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Event{}, id).Error
	})
}

func CountEventRegistrations(eventID uint) (int64, error) {
	var count int64
	err := config.DB.Model(&models.EventRegistration{}).Where("event_id = ?", eventID).Count(&count).Error
	return count, err
}
