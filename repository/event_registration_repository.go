package repository

import (
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/models"
)

func CreateEventRegistration(reg *models.EventRegistration) error {
	return config.DB.Create(reg).Error
}

func FindEventRegistration(eventID, userID uint) (*models.EventRegistration, error) {
	var reg models.EventRegistration
	err := config.DB.Where("event_id = ? AND user_id = ?", eventID, userID).First(&reg).Error
	if err != nil {
		return nil, err
	}
	return &reg, nil
}

func DeleteEventRegistration(eventID, userID uint) error {
	return config.DB.
		Where("event_id = ? AND user_id = ?", eventID, userID).
		Delete(&models.EventRegistration{}).Error
}

func FindEventParticipants(eventID uint) ([]models.EventRegistration, error) {
	var regs []models.EventRegistration
	err := config.DB.
		Preload("User").
		Where("event_id = ?", eventID).
		Find(&regs).Error
	return regs, err
}
