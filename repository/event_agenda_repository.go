package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type EventAgendaRepository interface {
	CreateEventAgenda(agenda *models.EventAgenda) error
	FindEventAgendaByID(id uint) (*models.EventAgenda, error)
	UpdateEventAgenda(agenda *models.EventAgenda) error
	DeleteEventAgenda(id uint) error
}

type eventAgendaRepository struct {
	db *gorm.DB
}

func NewEventAgendaRepository(db *gorm.DB) EventAgendaRepository {
	return &eventAgendaRepository{db: db}
}

func (r *eventAgendaRepository) CreateEventAgenda(agenda *models.EventAgenda) error {
	return r.db.Create(agenda).Error
}

func (r *eventAgendaRepository) FindEventAgendaByID(id uint) (*models.EventAgenda, error) {
	var agenda models.EventAgenda
	if err := r.db.First(&agenda, id).Error; err != nil {
		return nil, err
	}
	return &agenda, nil
}

func (r *eventAgendaRepository) UpdateEventAgenda(agenda *models.EventAgenda) error {
	return r.db.Save(agenda).Error
}

func (r *eventAgendaRepository) DeleteEventAgenda(id uint) error {
	return r.db.Delete(&models.EventAgenda{}, id).Error
}
