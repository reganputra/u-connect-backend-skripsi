package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type AdminRepository interface {
	// Dashboard
	GetStats() (map[string]int64, error)

	// User management
	FindUsers(page, limit int, role string) ([]models.User, int64, error)
	FindUserByID(id uint) (*models.User, error)
	UpdateUser(u *models.User) error

	// Direct content deletion
	DeletePost(id uint) error
	DeleteGroup(id uint) error
	DeleteEvent(id uint) error
	DeleteJob(id uint) error
}

type adminRepository struct{ db *gorm.DB }

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

func (r *adminRepository) GetStats() (map[string]int64, error) {
	stats := make(map[string]int64)
	counts := []struct {
		key   string
		model interface{}
		where string
	}{
		{"users", &models.User{}, ""},
		{"posts", &models.Post{}, ""},
		{"groups", &models.Group{}, ""},
		{"events", &models.Event{}, ""},
		{"jobs", &models.Job{}, ""},
		{"reports_pending", &models.Report{}, "status = 'pending'"},
	}
	for _, c := range counts {
		var n int64
		q := r.db.Model(c.model)
		if c.where != "" {
			q = q.Where(c.where)
		}
		if err := q.Count(&n).Error; err != nil {
			return nil, err
		}
		stats[c.key] = n
	}
	return stats, nil
}

// ─── User Management ──────────────────────────────────────────────────────────

func (r *adminRepository) FindUsers(page, limit int, role string) ([]models.User, int64, error) {
	var users []models.User
	var total int64
	offset := (page - 1) * limit

	q := r.db.Model(&models.User{})
	if role != "" {
		q = q.Where("role = ?", role)
	}
	q.Count(&total)
	err := q.Order("created_at desc").Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}

func (r *adminRepository) FindUserByID(id uint) (*models.User, error) {
	var u models.User
	if err := r.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *adminRepository) UpdateUser(u *models.User) error {
	return r.db.Save(u).Error
}

// ─── Direct Content Deletion ──────────────────────────────────────────────────

func (r *adminRepository) DeletePost(id uint) error {
	// Cascade: delete reactions, votes, comments then post
	r.db.Where("post_id = ?", id).Delete(&models.Reaction{})
	r.db.Where("post_id = ?", id).Delete(&models.Vote{})
	r.db.Where("post_id = ?", id).Delete(&models.Comment{})
	return r.db.Delete(&models.Post{}, id).Error
}

func (r *adminRepository) DeleteGroup(id uint) error {
	r.db.Where("article_id IN (SELECT id FROM group_articles WHERE group_id = ?)", id).Delete(&models.GroupReaction{})
	r.db.Where("comment_id IN (SELECT id FROM group_comments WHERE article_id IN (SELECT id FROM group_articles WHERE group_id = ?))", id).Delete(&models.GroupReaction{})
	r.db.Where("article_id IN (SELECT id FROM group_articles WHERE group_id = ?)", id).Delete(&models.GroupComment{})
	r.db.Where("group_id = ?", id).Delete(&models.GroupArticle{})
	r.db.Where("group_id = ?", id).Delete(&models.GroupMember{})
	return r.db.Delete(&models.Group{}, id).Error
}

func (r *adminRepository) DeleteEvent(id uint) error {
	r.db.Where("event_id = ?", id).Delete(&models.EventRegistration{})
	r.db.Where("event_id = ?", id).Delete(&models.EventAgenda{})
	return r.db.Delete(&models.Event{}, id).Error
}

func (r *adminRepository) DeleteJob(id uint) error {
	r.db.Where("job_id = ?", id).Delete(&models.JobApplication{})
	return r.db.Delete(&models.Job{}, id).Error
}
