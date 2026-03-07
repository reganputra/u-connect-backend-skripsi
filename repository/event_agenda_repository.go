package repository

import (
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/models"
)

func CreateEventAgenda(agenda *models.EventAgenda) error {
	return config.DB.Create(agenda).Error
}

func FindEventAgendaByID(id uint) (*models.EventAgenda, error) {
	var agenda models.EventAgenda
	err := config.DB.First(&agenda, id).Error
	if err != nil {
		return nil, err
	}
	return &agenda, nil
}

func UpdateEventAgenda(agenda *models.EventAgenda) error {
	return config.DB.Save(agenda).Error
}

func DeleteEventAgenda(id uint) error {
	return config.DB.Delete(&models.EventAgenda{}, id).Error
}
