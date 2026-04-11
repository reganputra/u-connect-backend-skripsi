package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
)

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
}

func NewAdminService(
	adminRepo repository.AdminRepository,
	reportRepo repository.ReportRepository,
	categoryRepo repository.CategoryRepository,
	notifSvc NotificationService,
) AdminService {
	return &adminService{
		adminRepo:    adminRepo,
		reportRepo:   reportRepo,
		categoryRepo: categoryRepo,
		notifSvc:     notifSvc,
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
		switch report.TargetType {
		case "post":
			_ = s.adminRepo.DeletePost(report.TargetID)
		case "group":
			_ = s.adminRepo.DeleteGroup(report.TargetID)
		case "event":
			_ = s.adminRepo.DeleteEvent(report.TargetID)
		case "job":
			_ = s.adminRepo.DeleteJob(report.TargetID)
		}
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
	// Notify the reporter
	_ = s.notifSvc.Notify(
		report.ReporterID,
		"report_rejected",
		"Laporanmu ditolak",
		fmt.Sprintf("Admin: %s", req.AdminNote),
		"report",
		report.ID,
	)

	return report, nil
}

// ─── Direct Content Deletion ──────────────────────────────────────────────────

func (s *adminService) DeletePost(id uint) error {
	return s.adminRepo.DeletePost(id)
}
func (s *adminService) DeleteGroup(id uint) error {
	return s.adminRepo.DeleteGroup(id)
}
func (s *adminService) DeleteEvent(id uint) error {
	return s.adminRepo.DeleteEvent(id)
}
func (s *adminService) DeleteJob(id uint) error {
	return s.adminRepo.DeleteJob(id)
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
	cat := &models.Category{Name: req.Name, Description: req.Description}
	if err := s.categoryRepo.Create(cat); err != nil {
		return nil, errors.New("gagal membuat kategori (nama mungkin sudah ada)")
	}
	return cat, nil
}

func (s *adminService) UpdateCategory(id uint, req CategoryRequest) (*models.Category, error) {
	cat, err := s.categoryRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("kategori tidak ditemukan")
	}
	if req.Name != "" {
		cat.Name = req.Name
	}
	if req.Description != nil {
		cat.Description = req.Description
	}
	if err := s.categoryRepo.Update(cat); err != nil {
		return nil, errors.New("gagal memperbarui kategori")
	}
	return cat, nil
}

func (s *adminService) DeleteCategory(id uint) error {
	if _, err := s.categoryRepo.FindByID(id); err != nil {
		return errors.New("kategori tidak ditemukan")
	}
	return s.categoryRepo.Delete(id)
}
