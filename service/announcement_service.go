package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
	"gorm.io/gorm"
)

type CreateAnnouncementRequest struct {
	Title      string     `json:"title"`
	Message    string     `json:"message"`
	ActionURL  *string    `json:"action_url"`
	ActionText *string    `json:"action_text"`
	TargetRole string     `json:"target_role"` // all | student | alumni | partner
	Priority   string     `json:"priority"`    // info | warning | urgent
	IsBanner   bool       `json:"is_banner"`
	ExpiresAt  *time.Time `json:"expires_at"`
}

type AnnouncementService interface {
	CreateBroadcast(adminID uint, req CreateAnnouncementRequest, ip, userAgent string) (*models.Announcement, error)
	GetActiveBanners(userRole string) ([]models.Announcement, error)
	GetAnnouncements(page, limit int) ([]models.Announcement, int64, error)
	DeleteAnnouncement(adminID uint, id uint, ip, userAgent string) error
}

type announcementService struct {
	repo     repository.AnnouncementRepository
	userRepo repository.UserRepository
	notifSvc NotificationService
	adminSvc AdminService
	db       *gorm.DB
}

func NewAnnouncementService(
	repo repository.AnnouncementRepository,
	userRepo repository.UserRepository,
	notifSvc NotificationService,
	adminSvc AdminService,
	db *gorm.DB,
) AnnouncementService {
	return &announcementService{
		repo:     repo,
		userRepo: userRepo,
		notifSvc: notifSvc,
		adminSvc: adminSvc,
		db:       db,
	}
}

func (s *announcementService) CreateBroadcast(adminID uint, req CreateAnnouncementRequest, ip, userAgent string) (*models.Announcement, error) {
	if req.Title == "" || req.Message == "" {
		return nil, errors.New("judul dan pesan pengumuman wajib diisi")
	}

	targetRole := req.TargetRole
	if targetRole == "" {
		targetRole = "all"
	}
	priority := req.Priority
	if priority == "" {
		priority = "info"
	}

	a := &models.Announcement{
		AdminID:    adminID,
		Title:      req.Title,
		Message:    req.Message,
		ActionURL:  req.ActionURL,
		ActionText: req.ActionText,
		TargetRole: targetRole,
		Priority:   priority,
		IsBanner:   req.IsBanner,
		ExpiresAt:  req.ExpiresAt,
	}

	if err := s.repo.Create(a); err != nil {
		return nil, errors.New("gagal membuat pengumuman")
	}

	// Async: Bulk notify target users & log audit
	go func() {
		var targetUsers []models.User
		query := s.db.Where("is_active = true")
		if targetRole != "all" {
			query = query.Where("role = ?", targetRole)
		}
		if err := query.Select("id").Find(&targetUsers).Error; err == nil {
			for _, u := range targetUsers {
				_ = s.notifSvc.Notify(
					u.ID,
					"broadcast",
					a.Title,
					a.Message,
					"announcement",
					a.ID,
				)
			}
		}

		detail := fmt.Sprintf("Membuat pengumuman broadcast ID %d (Judul: '%s', Target: %s, Prioritas: %s, Banner: %v)", a.ID, a.Title, a.TargetRole, a.Priority, a.IsBanner)
		_ = s.adminSvc.RecordActivityLog(adminID, "CREATE_BROADCAST", "announcement", a.ID, &detail, &ip, &userAgent)
	}()

	return a, nil
}

func (s *announcementService) GetActiveBanners(userRole string) ([]models.Announcement, error) {
	return s.repo.FindActiveBanners(userRole)
}

func (s *announcementService) GetAnnouncements(page, limit int) ([]models.Announcement, int64, error) {
	return s.repo.FindAll(page, limit)
}

func (s *announcementService) DeleteAnnouncement(adminID uint, id uint, ip, userAgent string) error {
	a, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("pengumuman tidak ditemukan")
	}

	if err := s.repo.Delete(id); err != nil {
		return errors.New("gagal menghapus pengumuman")
	}

	detail := fmt.Sprintf("Menghapus pengumuman broadcast ID %d (Judul: '%s')", id, a.Title)
	_ = s.adminSvc.RecordActivityLog(adminID, "DELETE_BROADCAST", "announcement", id, &detail, &ip, &userAgent)

	return nil
}
