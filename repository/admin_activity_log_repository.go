package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type AdminActivityLogRepository interface {
	CreateLog(log *models.AdminActivityLog) error
	FindLogs(page, limit int, actionFilter, targetTypeFilter string) ([]models.AdminActivityLog, int64, error)
}

type adminActivityLogRepository struct {
	db *gorm.DB
}

func NewAdminActivityLogRepository(db *gorm.DB) AdminActivityLogRepository {
	return &adminActivityLogRepository{db: db}
}

func (r *adminActivityLogRepository) CreateLog(log *models.AdminActivityLog) error {
	return r.db.Create(log).Error
}

func (r *adminActivityLogRepository) FindLogs(page, limit int, actionFilter, targetTypeFilter string) ([]models.AdminActivityLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var logs []models.AdminActivityLog
	var total int64

	query := r.db.Model(&models.AdminActivityLog{})

	if actionFilter != "" {
		query = query.Where("action = ?", actionFilter)
	}
	if targetTypeFilter != "" {
		query = query.Where("target_type = ?", targetTypeFilter)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("Admin").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error

	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
