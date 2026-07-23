package repository

import (
	"time"

	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type AnnouncementRepository interface {
	Create(a *models.Announcement) error
	FindAll(page, limit int) ([]models.Announcement, int64, error)
	FindActiveBanners(userRole string) ([]models.Announcement, error)
	FindByID(id uint) (*models.Announcement, error)
	Delete(id uint) error
}

type announcementRepository struct {
	db *gorm.DB
}

func NewAnnouncementRepository(db *gorm.DB) AnnouncementRepository {
	return &announcementRepository{db: db}
}

func (r *announcementRepository) Create(a *models.Announcement) error {
	return r.db.Create(a).Error
}

func (r *announcementRepository) FindAll(page, limit int) ([]models.Announcement, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var items []models.Announcement
	var total int64

	query := r.db.Model(&models.Announcement{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("Admin").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&items).Error

	return items, total, err
}

func (r *announcementRepository) FindActiveBanners(userRole string) ([]models.Announcement, error) {
	var items []models.Announcement
	now := time.Now()

	query := r.db.Model(&models.Announcement{}).
		Where("is_banner = true").
		Where("expires_at IS NULL OR expires_at > ?", now)

	if userRole != "" {
		query = query.Where("target_role = 'all' OR target_role = ?", userRole)
	}

	err := query.Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *announcementRepository) FindByID(id uint) (*models.Announcement, error) {
	var a models.Announcement
	err := r.db.Preload("Admin").First(&a, id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *announcementRepository) Delete(id uint) error {
	return r.db.Delete(&models.Announcement{}, id).Error
}
