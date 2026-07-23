package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/reganputra/skripsi-backend/constant"
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

type AdminReportView struct {
	models.Report
	TargetLabel       string `json:"TargetLabel"`
	TargetRedirectURL string `json:"TargetRedirectURL"`
	TargetExists      bool   `json:"TargetExists"`
}

// ─── Interface ────────────────────────────────────────────────────────────────

type AdminService interface {
	GetDashboardStats() (map[string]int64, error)

	// User management
	GetUsers(page, limit int, role, search string) ([]repository.AdminUserWithProfile, int64, error)
	GetUserByID(id uint) (*models.User, error)
	SetUserStatus(adminID uint, id uint, req UpdateUserStatusRequest, ip, userAgent string) (*models.User, error)
	SetUserRole(adminID uint, id uint, req UpdateUserRoleRequest, ip, userAgent string) (*models.User, error)

	// Report moderation
	GetReports(page, limit int, status string) ([]AdminReportView, int64, error)
	GetReportByID(id uint) (*AdminReportView, error)
	ResolveReport(adminID uint, reportID uint, req ResolveReportRequest, ip, userAgent string) (*models.Report, error)
	RejectReport(adminID uint, reportID uint, req RejectReportRequest, ip, userAgent string) (*models.Report, error)

	// Direct content deletion
	DeletePost(adminID uint, id uint, ip, userAgent string) error
	DeleteGroup(adminID uint, id uint, ip, userAgent string) error
	DeleteEvent(adminID uint, id uint, ip, userAgent string) error
	DeleteJob(adminID uint, id uint, ip, userAgent string) error

	// Category management
	GetCategories() ([]models.Category, error)
	CreateCategory(req CategoryRequest) (*models.Category, error)
	UpdateCategory(id uint, req CategoryRequest) (*models.Category, error)
	DeleteCategory(id uint) error

	// Admin activity logs
	GetAdminActivityLogs(page, limit int, action, targetType string) ([]models.AdminActivityLog, int64, error)
	RecordActivityLog(adminID uint, action, targetType string, targetID uint, details *string, ip, userAgent *string) error
}

// ─── Implementation ───────────────────────────────────────────────────────────

type adminService struct {
	adminRepo    repository.AdminRepository
	reportRepo   repository.ReportRepository
	categoryRepo repository.CategoryRepository
	logRepo      repository.AdminActivityLogRepository
	notifSvc     NotificationService
	db           *gorm.DB
}

func NewAdminService(
	adminRepo repository.AdminRepository,
	reportRepo repository.ReportRepository,
	categoryRepo repository.CategoryRepository,
	logRepo repository.AdminActivityLogRepository,
	notifSvc NotificationService,
	db *gorm.DB,
) AdminService {
	return &adminService{
		adminRepo:    adminRepo,
		reportRepo:   reportRepo,
		categoryRepo: categoryRepo,
		logRepo:      logRepo,
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

func (s *adminService) GetUsers(page, limit int, role, search string) ([]repository.AdminUserWithProfile, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.adminRepo.FindUsers(page, limit, role, search)
}

func (s *adminService) GetUserByID(id uint) (*models.User, error) {
	u, err := s.adminRepo.FindUserByID(id)
	if err != nil {
		return nil, errors.New("pengguna tidak ditemukan")
	}
	return u, nil
}

func (s *adminService) SetUserStatus(adminID uint, id uint, req UpdateUserStatusRequest, ip, userAgent string) (*models.User, error) {
	u, err := s.adminRepo.FindUserByID(id)
	if err != nil {
		return nil, errors.New("pengguna tidak ditemukan")
	}
	u.IsActive = req.IsActive
	if err := s.adminRepo.UpdateUser(u); err != nil {
		return nil, errors.New("gagal memperbarui status pengguna")
	}

	statusStr := "Aktif"
	if !u.IsActive {
		statusStr = "Non-Aktif (Banned)"
	}
	detail := fmt.Sprintf("Mengubah status akun user ID %d (Nama: %s) menjadi %s", id, u.Name, statusStr)
	_ = s.RecordActivityLog(adminID, "UPDATE_USER_STATUS", "user", id, &detail, &ip, &userAgent)

	return u, nil
}

func (s *adminService) SetUserRole(adminID uint, id uint, req UpdateUserRoleRequest, ip, userAgent string) (*models.User, error) {
	validRoles := map[string]bool{constant.RoleAlumni: true, constant.RoleStudent: true, constant.RolePartner: true, constant.RoleAdmin: true}
	if !validRoles[req.Role] {
		return nil, errors.New("role tidak valid: harus alumni, student, partner, atau admin")
	}
	u, err := s.adminRepo.FindUserByID(id)
	if err != nil {
		return nil, errors.New("pengguna tidak ditemukan")
	}
	oldRole := u.Role
	u.Role = req.Role
	if err := s.adminRepo.UpdateUser(u); err != nil {
		return nil, errors.New("gagal memperbarui role pengguna")
	}

	detail := fmt.Sprintf("Mengubah role user ID %d (Nama: %s) dari '%s' menjadi '%s'", id, u.Name, oldRole, u.Role)
	_ = s.RecordActivityLog(adminID, "UPDATE_USER_ROLE", "user", id, &detail, &ip, &userAgent)

	return u, nil
}

// ─── Report Moderation ────────────────────────────────────────────────────────

func (s *adminService) GetReports(page, limit int, status string) ([]AdminReportView, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	reports, total, err := s.reportRepo.FindReports(page, limit, status)
	if err != nil {
		return nil, 0, err
	}

	views := make([]AdminReportView, 0, len(reports))
	for i := range reports {
		views = append(views, s.buildAdminReportView(&reports[i]))
	}

	return views, total, nil
}

func (s *adminService) GetReportByID(id uint) (*AdminReportView, error) {
	r, err := s.reportRepo.FindReportByID(id)
	if err != nil {
		return nil, errors.New("laporan tidak ditemukan")
	}
	view := s.buildAdminReportView(r)
	return &view, nil
}

func (s *adminService) ResolveReport(adminID uint, reportID uint, req ResolveReportRequest, ip, userAgent string) (*models.Report, error) {
	report, err := s.reportRepo.FindReportByID(reportID)
	if err != nil {
		return nil, errors.New("laporan tidak ditemukan")
	}
	if report.Status != constant.StatusPending {
		return nil, errors.New("laporan sudah diproses sebelumnya")
	}

	targetLabel := fmt.Sprintf("konten (%s)", report.TargetType)
	if label, err := s.describeTargetForNotification(report.TargetType, report.TargetID); err == nil {
		targetLabel = label
	}

	contentDeleted := false
	// Optionally delete the reported content
	if req.DeleteContent {
		if err := s.deleteTargetAndNotify(adminID, report.TargetType, report.TargetID, req.AdminNote, ip, userAgent); err == nil {
			contentDeleted = true
		}
	}

	now := time.Now()
	report.Status = constant.StatusResolved
	report.AdminNote = req.AdminNote
	report.ResolvedByID = &adminID
	report.ResolvedAt = &now

	if err := s.reportRepo.UpdateReport(report); err != nil {
		return nil, errors.New("gagal memperbarui laporan")
	}

	if contentDeleted {
		reason := "Melanggar kebijakan komunitas"
		if req.AdminNote != nil && *req.AdminNote != "" {
			reason = *req.AdminNote
		}
		_ = s.notifSvc.Notify(
			report.ReporterID,
			"report_resolved_deleted",
			"Laporanmu ditindaklanjuti",
			fmt.Sprintf("Laporanmu terbukti valid. %s telah dihapus admin. Alasan: %s", targetLabel, reason),
			report.TargetType,
			report.TargetID,
		)
	}

	detail := fmt.Sprintf("Menyelesaikan laporan ID %d (Target: %s #%d)", reportID, report.TargetType, report.TargetID)
	if req.AdminNote != nil && *req.AdminNote != "" {
		detail += fmt.Sprintf(" - Catatan: %s", *req.AdminNote)
	}
	_ = s.RecordActivityLog(adminID, "RESOLVE_REPORT", "report", reportID, &detail, &ip, &userAgent)

	return report, nil
}

func (s *adminService) RejectReport(adminID uint, reportID uint, req RejectReportRequest, ip, userAgent string) (*models.Report, error) {
	if req.AdminNote == "" {
		return nil, errors.New("alasan penolakan wajib diisi")
	}
	report, err := s.reportRepo.FindReportByID(reportID)
	if err != nil {
		return nil, errors.New("laporan tidak ditemukan")
	}
	if report.Status != constant.StatusPending {
		return nil, errors.New("laporan sudah diproses sebelumnya")
	}

	now := time.Now()
	note := req.AdminNote
	report.Status = constant.StatusRejected
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

	detail := fmt.Sprintf("Menolak laporan ID %d (Target: %s #%d) - Alasan: %s", reportID, report.TargetType, report.TargetID, req.AdminNote)
	_ = s.RecordActivityLog(adminID, "REJECT_REPORT", "report", reportID, &detail, &ip, &userAgent)

	return report, nil
}

// ─── Direct Content Deletion ──────────────────────────────────────────────────

func (s *adminService) DeletePost(adminID uint, id uint, ip, userAgent string) error {
	return s.deleteTargetAndNotify(adminID, "post", id, nil, ip, userAgent)
}
func (s *adminService) DeleteGroup(adminID uint, id uint, ip, userAgent string) error {
	return s.deleteTargetAndNotify(adminID, "group", id, nil, ip, userAgent)
}
func (s *adminService) DeleteEvent(adminID uint, id uint, ip, userAgent string) error {
	return s.deleteTargetAndNotify(adminID, "event", id, nil, ip, userAgent)
}
func (s *adminService) DeleteJob(adminID uint, id uint, ip, userAgent string) error {
	return s.deleteTargetAndNotify(adminID, "job", id, nil, ip, userAgent)
}

func (s *adminService) deleteTargetAndNotify(adminID uint, targetType string, targetID uint, adminNote *string, ip, userAgent string) error {
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

	detail := fmt.Sprintf("Menghapus konten %s ID %d", targetType, targetID)
	_ = s.RecordActivityLog(adminID, "DELETE_CONTENT", targetType, targetID, &detail, &ip, &userAgent)

	return nil
}

// ─── Activity Logs ────────────────────────────────────────────────────────────

func (s *adminService) RecordActivityLog(adminID uint, action, targetType string, targetID uint, details *string, ip, userAgent *string) error {
	if adminID == 0 {
		return nil
	}
	log := &models.AdminActivityLog{
		AdminID:    adminID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Details:    details,
		IPAddress:  ip,
		UserAgent:  userAgent,
	}
	return s.logRepo.CreateLog(log)
}

func (s *adminService) GetAdminActivityLogs(page, limit int, actionFilter, targetTypeFilter string) ([]models.AdminActivityLog, int64, error) {
	return s.logRepo.FindLogs(page, limit, actionFilter, targetTypeFilter)
}

func (s *adminService) buildAdminReportView(report *models.Report) AdminReportView {
	label, redirectURL, exists := s.resolveReportTargetMeta(report.TargetType, report.TargetID)
	return AdminReportView{
		Report:            *report,
		TargetLabel:       label,
		TargetRedirectURL: redirectURL,
		TargetExists:      exists,
	}
}

func (s *adminService) resolveReportTargetMeta(targetType string, targetID uint) (string, string, bool) {
	switch targetType {
	case "post":
		var post models.Post
		if err := s.db.Select("id", "title").First(&post, targetID).Error; err != nil {
			return s.missingTargetMeta(targetType, targetID, err)
		}
		return fmt.Sprintf("Post \"%s\"", TruncateText(post.Title, 80)), fmt.Sprintf("/feed/%d", post.ID), true
	case "comment":
		var comment models.Comment
		if err := s.db.Select("id", "content", "post_id").First(&comment, targetID).Error; err != nil {
			return s.missingTargetMeta(targetType, targetID, err)
		}
		return fmt.Sprintf("Komentar \"%s\"", TruncateText(comment.Content, 80)), fmt.Sprintf("/feed/%d", comment.PostID), true
	case "group":
		var group models.Group
		if err := s.db.Select("id", "title").First(&group, targetID).Error; err != nil {
			return s.missingTargetMeta(targetType, targetID, err)
		}
		return fmt.Sprintf("Grup \"%s\"", TruncateText(group.Title, 80)), fmt.Sprintf("/groups/%d", group.ID), true
	case "group_article":
		var article models.GroupArticle
		if err := s.db.Select("id", "title", "group_id").First(&article, targetID).Error; err != nil {
			return s.missingTargetMeta(targetType, targetID, err)
		}
		return fmt.Sprintf("Artikel grup \"%s\"", TruncateText(article.Title, 80)), fmt.Sprintf("/groups/%d/article/%d", article.GroupID, article.ID), true
	case "event":
		var event models.Event
		if err := s.db.Select("id", "title").First(&event, targetID).Error; err != nil {
			return s.missingTargetMeta(targetType, targetID, err)
		}
		return fmt.Sprintf("Event \"%s\"", TruncateText(event.Title, 80)), fmt.Sprintf("/events/%d", event.ID), true
	case "job":
		var job models.Job
		if err := s.db.Select("id", "title").First(&job, targetID).Error; err != nil {
			return s.missingTargetMeta(targetType, targetID, err)
		}
		return fmt.Sprintf("Lowongan \"%s\"", TruncateText(job.Title, 80)), fmt.Sprintf("/jobs/%d", job.ID), true
	default:
		return fmt.Sprintf("Konten (%s) ID %d", targetType, targetID), "", false
	}
}

func (s *adminService) missingTargetMeta(targetType string, targetID uint, err error) (string, string, bool) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Sprintf("Konten (%s) ID %d [sudah dihapus]", targetType, targetID), "", false
	}
	return fmt.Sprintf("Konten (%s) ID %d", targetType, targetID), "", false
}

func (s *adminService) describeTargetForNotification(targetType string, targetID uint) (string, error) {
	label, _, exists := s.resolveReportTargetMeta(targetType, targetID)
	if !exists {
		return "", gorm.ErrRecordNotFound
	}
	return label, nil
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
