package repository

import (
	"time"

	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(n *models.Notification) error
	UpdateRedirectURL(id uint, url string) error
	GetByUserID(userID uint, page, limit int) ([]models.Notification, int64, error)
	MarkAsRead(notificationID, userID uint) error
	MarkAllAsRead(userID uint) error
	CountUnread(userID uint) (int64, error)
	ExistsRecent(userID uint, notifType, refType string, refID uint, within time.Duration) (bool, error)
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(n *models.Notification) error {
	return r.db.Create(n).Error
}

func (r *notificationRepository) UpdateRedirectURL(id uint, url string) error {
	return r.db.Model(&models.Notification{}).Where("id = ?", id).Update("redirect_url", url).Error
}

func (r *notificationRepository) GetByUserID(userID uint, page, limit int) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	base := r.db.Model(&models.Notification{}).Where("user_id = ?", userID)
	base.Count(&total)

	offset := (page - 1) * limit
	err := base.Order("created_at DESC").Offset(offset).Limit(limit).Find(&notifications).Error
	return notifications, total, err
}

func (r *notificationRepository) MarkAsRead(notificationID, userID uint) error {
	result := r.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("is_read", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *notificationRepository) MarkAllAsRead(userID uint) error {
	return r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).Error
}

func (r *notificationRepository) CountUnread(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&count).Error
	return count, err
}

func (r *notificationRepository) ExistsRecent(userID uint, notifType, refType string, refID uint, within time.Duration) (bool, error) {
	var count int64
	since := time.Now().Add(-within)
	err := r.db.Model(&models.Notification{}).
		Where("user_id = ? AND notification_type = ? AND reference_type = ? AND reference_id = ? AND created_at >= ?",
			userID, notifType, refType, refID, since).
		Count(&count).Error
	return count > 0, err
}
