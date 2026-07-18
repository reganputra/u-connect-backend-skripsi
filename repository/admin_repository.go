package repository

import (
	"strings"
	"time"

	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type AdminRepository interface {
	// Dashboard
	GetStats() (map[string]int64, error)

	// User management
	FindUsers(page, limit int, role, search string) ([]AdminUserWithProfile, int64, error)
	FindUserByID(id uint) (*models.User, error)
	UpdateUser(u *models.User) error

	// Direct content deletion
	DeletePost(id uint) error
	DeleteGroup(id uint) error
	DeleteEvent(id uint) error
	DeleteJob(id uint) error
}

// AdminUserWithProfile is the flat DTO returned by the admin user list.
// It joins User + UserProfile so the frontend can display AND edit profile
// fields inline without a separate profile request.
type AdminUserWithProfile struct {
	// User fields
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	// Profile fields (nullable — user may not have a profile yet)
	Name              *string `json:"name"`
	ProfilePicture    *string `json:"profile_picture"`
	Bio               *string `json:"bio"`
	Skills            *string `json:"skills"`
	Interests         *string `json:"interests"`
	Position          *string `json:"position"`
	IndustryName      *string `json:"industry_name"`
	MentorQuota       *int    `json:"mentor_quota"`
	MentorDescription *string `json:"mentor_description"`
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

func (r *adminRepository) FindUsers(page, limit int, role, search string) ([]AdminUserWithProfile, int64, error) {
	var results []AdminUserWithProfile
	var total int64
	offset := (page - 1) * limit

	q := r.db.Table("users u").
		Select(`u.id, u.email, u.role, u.is_active, u.created_at, u.name,
			   up.profile_picture, up.bio,
			   up.skills, up.interests, up.position,
			   up.industry_name, up.mentor_quota, up.mentor_description`).
		Joins("LEFT JOIN user_profiles up ON up.user_id = u.id")

	if role != "" {
		q = q.Where("u.role = ?", role)
	}
	if s := strings.TrimSpace(search); s != "" {
		like := "%" + strings.ToLower(s) + "%"
		q = q.Where("LOWER(u.email) LIKE ? OR LOWER(u.name) LIKE ?", like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("u.created_at DESC").Offset(offset).Limit(limit).Scan(&results).Error
	return results, total, err
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
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", id).Delete(&models.Reaction{}).Error; err != nil {
			return err
		}
		if err := tx.Where("post_id = ?", id).Delete(&models.Vote{}).Error; err != nil {
			return err
		}
		if err := tx.Where("post_id = ?", id).Delete(&models.Comment{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Post{}, id).Error
	})
}

func (r *adminRepository) DeleteGroup(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("article_id IN (SELECT id FROM group_articles WHERE group_id = ?)", id).Delete(&models.GroupReaction{}).Error; err != nil {
			return err
		}
		if err := tx.Where("comment_id IN (SELECT id FROM group_comments WHERE article_id IN (SELECT id FROM group_articles WHERE group_id = ?))", id).Delete(&models.GroupReaction{}).Error; err != nil {
			return err
		}
		if err := tx.Where("article_id IN (SELECT id FROM group_articles WHERE group_id = ?)", id).Delete(&models.GroupComment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("group_id = ?", id).Delete(&models.GroupArticle{}).Error; err != nil {
			return err
		}
		if err := tx.Where("group_id = ?", id).Delete(&models.GroupMember{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Group{}, id).Error
	})
}

func (r *adminRepository) DeleteEvent(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("event_id = ?", id).Delete(&models.EventRegistration{}).Error; err != nil {
			return err
		}
		if err := tx.Where("event_id = ?", id).Delete(&models.EventAgenda{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Event{}, id).Error
	})
}

func (r *adminRepository) DeleteJob(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("job_id = ?", id).Delete(&models.JobApplication{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Job{}, id).Error
	})
}
