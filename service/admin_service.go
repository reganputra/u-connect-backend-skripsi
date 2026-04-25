package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
	"gorm.io/gorm"
)

const generalCategoryName = "General"

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type UpdateUserStatusRequest struct {
	IsActive bool `json:"is_active"`
}

type UpdateUserRoleRequest struct {
	Role string `json:"role"`
}

type ResolveReportRequest struct {
	AdminNote     *string `json:"admin_note"`
	DeleteContent bool    `json:"delete_content"` // if true, delete the reported content
}

type RejectReportRequest struct {
	AdminNote string `json:"admin_note"`
}

type CategoryRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// ─── Interface ────────────────────────────────────────────────────────────────

type AdminService interface {
	GetDashboardStats() (map[string]int64, error)

	// User management
	GetUsers(page, limit int, role string) ([]models.User, int64, error)
	GetUserByID(id uint) (*models.User, error)
	SetUserStatus(id uint, req UpdateUserStatusRequest) (*models.User, error)
	SetUserRole(id uint, req UpdateUserRoleRequest) (*models.User, error)

	// Report moderation
	GetReports(page, limit int, status string) ([]models.Report, int64, error)
	GetReportByID(id uint) (*models.Report, error)
	ResolveReport(adminID uint, reportID uint, req ResolveReportRequest) (*models.Report, error)
	RejectReport(adminID uint, reportID uint, req RejectReportRequest) (*models.Report, error)

	// Direct content deletion
	DeletePost(id uint) error
	DeleteGroup(id uint) error
	DeleteEvent(id uint) error
	DeleteJob(id uint) error

	// Category management
	GetCategories() ([]models.Category, error)
	CreateCategory(req CategoryRequest) (*models.Category, error)
	UpdateCategory(id uint, req CategoryRequest) (*models.Category, error)
	DeleteCategory(id uint) error
}

// ─── Implementation ───────────────────────────────────────────────────────────

type adminService struct {
	adminRepo    repository.AdminRepository
	reportRepo   repository.ReportRepository
	categoryRepo repository.CategoryRepository
	notifSvc     NotificationService
	db           *gorm.DB
}

func NewAdminService(
	adminRepo repository.AdminRepository,
	reportRepo repository.ReportRepository,
	categoryRepo repository.CategoryRepository,
	notifSvc NotificationService,
	db *gorm.DB,
) AdminService {
	return &adminService{
		adminRepo:    adminRepo,
		reportRepo:   reportRepo,
		categoryRepo: categoryRepo,
		notifSvc:     notifSvc,
		db:           db,
	}
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

func (s *adminService) GetDashboardStats() (map[string]int64, error) {
	stats, err := s.adminRepo.GetStats()
	if err != nil {
		return nil, errors.New("gagal mengambil statistik dashboard")
	}
	return stats, nil
}

// ─── User Management ──────────────────────────────────────────────────────────

func (s *adminService) GetUsers(page, limit int, role string) ([]models.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.adminRepo.FindUsers(page, limit, role)
}

func (s *adminService) GetUserByID(id uint) (*models.User, error) {
	u, err := s.adminRepo.FindUserByID(id)
	if err != nil {
		return nil, errors.New("pengguna tidak ditemukan")
	}
	return u, nil
}

func (s *adminService) SetUserStatus(id uint, req UpdateUserStatusRequest) (*models.User, error) {
	u, err := s.adminRepo.FindUserByID(id)
	if err != nil {
		return nil, errors.New("pengguna tidak ditemukan")
	}
	u.IsActive = req.IsActive
	if err := s.adminRepo.UpdateUser(u); err != nil {
		return nil, errors.New("gagal memperbarui status pengguna")
	}
	return u, nil
}

func (s *adminService) SetUserRole(id uint, req UpdateUserRoleRequest) (*models.User, error) {
	validRoles := map[string]bool{"alumni": true, "student": true, "partner": true, "admin": true}
	if !validRoles[req.Role] {
		return nil, errors.New("role tidak valid: harus alumni, student, partner, atau admin")
	}
	u, err := s.adminRepo.FindUserByID(id)
	if err != nil {
		return nil, errors.New("pengguna tidak ditemukan")
	}
	u.Role = req.Role
	if err := s.adminRepo.UpdateUser(u); err != nil {
		return nil, errors.New("gagal memperbarui role pengguna")
	}
	return u, nil
}

// ─── Report Moderation ────────────────────────────────────────────────────────

func (s *adminService) GetReports(page, limit int, status string) ([]models.Report, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.reportRepo.FindReports(page, limit, status)
}

func (s *adminService) GetReportByID(id uint) (*models.Report, error) {
	r, err := s.reportRepo.FindReportByID(id)
	if err != nil {
		return nil, errors.New("laporan tidak ditemukan")
	}
	return r, nil
}

func (s *adminService) ResolveReport(adminID uint, reportID uint, req ResolveReportRequest) (*models.Report, error) {
	report, err := s.reportRepo.FindReportByID(reportID)
	if err != nil {
		return nil, errors.New("laporan tidak ditemukan")
	}
	if report.Status != "pending" {
		return nil, errors.New("laporan sudah diproses sebelumnya")
	}

	// Optionally delete the reported content
	if req.DeleteContent {
		_ = s.deleteTargetAndNotify(report.TargetType, report.TargetID, req.AdminNote)
	}

	now := time.Now()
	report.Status = "resolved"
	report.AdminNote = req.AdminNote
	report.ResolvedByID = &adminID
	report.ResolvedAt = &now

	if err := s.reportRepo.UpdateReport(report); err != nil {
		return nil, errors.New("gagal memperbarui laporan")
	}
	return report, nil
}

func (s *adminService) RejectReport(adminID uint, reportID uint, req RejectReportRequest) (*models.Report, error) {
	if req.AdminNote == "" {
		return nil, errors.New("alasan penolakan wajib diisi")
	}
	report, err := s.reportRepo.FindReportByID(reportID)
	if err != nil {
		return nil, errors.New("laporan tidak ditemukan")
	}
	if report.Status != "pending" {
		return nil, errors.New("laporan sudah diproses sebelumnya")
	}

	now := time.Now()
	note := req.AdminNote
	report.Status = "rejected"
	report.AdminNote = &note
	report.ResolvedByID = &adminID
	report.ResolvedAt = &now

	if err := s.reportRepo.UpdateReport(report); err != nil {
		return nil, errors.New("gagal memperbarui laporan")
	}
	targetLabel, err := s.describeTargetForNotification(report.TargetType, report.TargetID)
	if err != nil {
		targetLabel = fmt.Sprintf("konten (%s)", report.TargetType)
	}
	// Notify the reporter
	_ = s.notifSvc.Notify(
		report.ReporterID,
		"report_rejected",
		"Laporanmu ditolak",
		fmt.Sprintf("Laporanmu terkait %s ditolak. Alasan: %s", targetLabel, req.AdminNote),
		report.TargetType,
		report.TargetID,
	)

	return report, nil
}

// ─── Direct Content Deletion ──────────────────────────────────────────────────

func (s *adminService) DeletePost(id uint) error {
	return s.deleteTargetAndNotify("post", id, nil)
}
func (s *adminService) DeleteGroup(id uint) error {
	return s.deleteTargetAndNotify("group", id, nil)
}
func (s *adminService) DeleteEvent(id uint) error {
	return s.deleteTargetAndNotify("event", id, nil)
}
func (s *adminService) DeleteJob(id uint) error {
	return s.deleteTargetAndNotify("job", id, nil)
}

func (s *adminService) deleteTargetAndNotify(targetType string, targetID uint, adminNote *string) error {
	ownerID, err := s.findContentOwner(targetType, targetID)
	if err != nil {
		return err
	}

	targetLabel, err := s.describeTargetForNotification(targetType, targetID)
	if err != nil {
		targetLabel = fmt.Sprintf("konten (%s)", targetType)
	}

	if err := s.deleteContentTarget(targetType, targetID); err != nil {
		return err
	}

	reason := "Melanggar kebijakan komunitas"
	if adminNote != nil && *adminNote != "" {
		reason = *adminNote
	}
	body := fmt.Sprintf("%s dihapus oleh admin. Alasan: %s", targetLabel, reason)

	_ = s.notifSvc.Notify(
		ownerID,
		"content_deleted_by_admin",
		"Konten dihapus admin",
		body,
		targetType,
		targetID,
	)

	return nil
}

func (s *adminService) describeTargetForNotification(targetType string, targetID uint) (string, error) {
	switch targetType {
	case "post":
		var post models.Post
		if err := s.db.Select("id", "title").First(&post, targetID).Error; err != nil {
			return "", err
		}
		return fmt.Sprintf("Post \"%s\"", truncateNotifyText(post.Title, 80)), nil
	case "group":
		var group models.Group
		if err := s.db.Select("id", "title").First(&group, targetID).Error; err != nil {
			return "", err
		}
		return fmt.Sprintf("Grup \"%s\"", truncateNotifyText(group.Title, 80)), nil
	case "event":
		var event models.Event
		if err := s.db.Select("id", "title").First(&event, targetID).Error; err != nil {
			return "", err
		}
		return fmt.Sprintf("Event \"%s\"", truncateNotifyText(event.Title, 80)), nil
	case "job":
		var job models.Job
		if err := s.db.Select("id", "title").First(&job, targetID).Error; err != nil {
			return "", err
		}
		return fmt.Sprintf("Lowongan \"%s\"", truncateNotifyText(job.Title, 80)), nil
	case "comment":
		var comment models.Comment
		if err := s.db.Select("id", "content").First(&comment, targetID).Error; err != nil {
			return "", err
		}
		return fmt.Sprintf("Komentar \"%s\"", truncateNotifyText(comment.Content, 80)), nil
	case "group_article":
		var article models.GroupArticle
		if err := s.db.Select("id", "title").First(&article, targetID).Error; err != nil {
			return "", err
		}
		return fmt.Sprintf("Artikel grup \"%s\"", truncateNotifyText(article.Title, 80)), nil
	default:
		return fmt.Sprintf("Konten (%s) ID %d", targetType, targetID), nil
	}
}

func truncateNotifyText(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 3 {
		return string(runes[:max])
	}
	return string(runes[:max-3]) + "..."
}

func (s *adminService) findContentOwner(targetType string, targetID uint) (uint, error) {
	switch targetType {
	case "post":
		var post models.Post
		if err := s.db.Select("id", "user_id").First(&post, targetID).Error; err != nil {
			return 0, err
		}
		return post.UserID, nil
	case "group":
		var group models.Group
		if err := s.db.Select("id", "owner_id").First(&group, targetID).Error; err != nil {
			return 0, err
		}
		return group.OwnerID, nil
	case "event":
		var event models.Event
		if err := s.db.Select("id", "user_id").First(&event, targetID).Error; err != nil {
			return 0, err
		}
		return event.UserID, nil
	case "job":
		var job models.Job
		if err := s.db.Select("id", "user_id").First(&job, targetID).Error; err != nil {
			return 0, err
		}
		return job.UserID, nil
	case "comment":
		var comment models.Comment
		if err := s.db.Select("id", "user_id").First(&comment, targetID).Error; err != nil {
			return 0, err
		}
		return comment.UserID, nil
	case "group_article":
		var article models.GroupArticle
		if err := s.db.Select("id", "user_id").First(&article, targetID).Error; err != nil {
			return 0, err
		}
		return article.UserID, nil
	default:
		return 0, errors.New("tipe target laporan tidak didukung")
	}
}

func (s *adminService) deleteContentTarget(targetType string, targetID uint) error {
	switch targetType {
	case "post":
		return s.adminRepo.DeletePost(targetID)
	case "group":
		return s.adminRepo.DeleteGroup(targetID)
	case "event":
		return s.adminRepo.DeleteEvent(targetID)
	case "job":
		return s.adminRepo.DeleteJob(targetID)
	case "comment":
		s.db.Where("comment_id = ?", targetID).Delete(&models.Reaction{})
		s.db.Where("comment_id = ?", targetID).Delete(&models.Vote{})
		return s.db.Delete(&models.Comment{}, targetID).Error
	case "group_article":
		s.db.Where("article_id = ?", targetID).Delete(&models.GroupReaction{})
		s.db.Where("comment_id IN (SELECT id FROM group_comments WHERE article_id = ?)", targetID).Delete(&models.GroupReaction{})
		s.db.Where("article_id = ?", targetID).Delete(&models.GroupComment{})
		s.db.Where("article_id = ?", targetID).Delete(&models.GroupArticleImage{})
		return s.db.Delete(&models.GroupArticle{}, targetID).Error
	default:
		return errors.New("tipe target laporan tidak didukung")
	}
}

// ─── Categories ───────────────────────────────────────────────────────────────

func (s *adminService) GetCategories() ([]models.Category, error) {
	cats, err := s.categoryRepo.FindAll()
	if err != nil {
		return nil, errors.New("gagal mengambil daftar kategori")
	}
	return cats, nil
}

func (s *adminService) CreateCategory(req CategoryRequest) (*models.Category, error) {
	if req.Name == "" {
		return nil, errors.New("nama kategori wajib diisi")
	}
	if req.Name == generalCategoryName {
		return nil, errors.New("kategori General sudah tersedia sebagai kategori cadangan")
	}
	cat := &models.Category{Name: req.Name, Description: req.Description}
	if err := s.categoryRepo.Create(cat); err != nil {
		return nil, errors.New("gagal membuat kategori (nama mungkin sudah ada)")
	}
	return cat, nil
}

func (s *adminService) UpdateCategory(id uint, req CategoryRequest) (*models.Category, error) {
	var updated *models.Category
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var cat models.Category
		if err := tx.First(&cat, id).Error; err != nil {
			return errors.New("kategori tidak ditemukan")
		}

		oldName := cat.Name
		if req.Name != "" {
			cat.Name = req.Name
		}
		if req.Description != nil {
			cat.Description = req.Description
		}

		if err := tx.Save(&cat).Error; err != nil {
			return errors.New("gagal memperbarui kategori")
		}

		if req.Name != "" && req.Name != oldName {
			if err := s.replaceCategoryInContent(tx, oldName, req.Name); err != nil {
				return err
			}
		}

		updated = &cat
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *adminService) DeleteCategory(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var cat models.Category
		if err := tx.First(&cat, id).Error; err != nil {
			return errors.New("kategori tidak ditemukan")
		}
		if cat.Name == generalCategoryName {
			return errors.New("kategori General tidak dapat dihapus")
		}

		if err := s.ensureGeneralCategory(tx); err != nil {
			return err
		}
		if err := s.replaceCategoryInContent(tx, cat.Name, generalCategoryName); err != nil {
			return err
		}
		if err := tx.Delete(&models.Category{}, id).Error; err != nil {
			return errors.New("gagal menghapus kategori")
		}
		return nil
	})
}

func (s *adminService) ensureGeneralCategory(tx *gorm.DB) error {
	var cat models.Category
	if err := tx.Where("name = ?", generalCategoryName).First(&cat).Error; err == nil {
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("gagal menyiapkan kategori General")
	}

	return tx.Create(&models.Category{Name: generalCategoryName}).Error
}

func (s *adminService) replaceCategoryInContent(tx *gorm.DB, oldName, newName string) error {
	if err := tx.Model(&models.Post{}).
		Where("category = ?", oldName).
		Update("category", newName).Error; err != nil {
		return errors.New("gagal memperbarui kategori pada postingan")
	}

	if err := tx.Model(&models.Group{}).
		Where("category = ?", oldName).
		Update("category", newName).Error; err != nil {
		return errors.New("gagal memperbarui kategori pada grup")
	}

	if err := tx.Model(&models.PortfolioItem{}).
		Where("category = ?", oldName).
		Update("category", newName).Error; err != nil {
		return errors.New("gagal memperbarui kategori pada portofolio")
	}

	return nil
}
