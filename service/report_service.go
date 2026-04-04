package service

import (
	"errors"

	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
)

var validTargetTypes = map[string]bool{
	"post":          true,
	"group":         true,
	"event":         true,
	"job":           true,
	"comment":       true,
	"group_article": true,
}

var validReportTypes = map[string]bool{
	"harassment":     true,
	"violence":       true,
	"hate_speech":    true,
	"spam":           true,
	"inappropriate":  true,
	"misinformation": true,
	"copyright":      true,
	"other":          true,
}

// ─── DTO ─────────────────────────────────────────────────────────────────────

type ReportRequest struct {
	TargetType  string  `json:"target_type"`
	TargetID    uint    `json:"target_id"`
	ReportType  string  `json:"report_type"`
	Description *string `json:"description"`
}

// ─── Interface ────────────────────────────────────────────────────────────────

type ReportService interface {
	CreateReport(reporterID uint, req ReportRequest) (*models.Report, error)
	GetMyReports(reporterID uint, page, limit int) ([]models.Report, int64, error)
}

// ─── Implementation ───────────────────────────────────────────────────────────

type reportService struct {
	reportRepo repository.ReportRepository
}

func NewReportService(reportRepo repository.ReportRepository) ReportService {
	return &reportService{reportRepo: reportRepo}
}

func (s *reportService) CreateReport(reporterID uint, req ReportRequest) (*models.Report, error) {
	if req.TargetType == "" {
		return nil, errors.New("target_type wajib diisi")
	}
	if !validTargetTypes[req.TargetType] {
		return nil, errors.New("target_type tidak valid: harus post, group, event, job, comment, atau group_article")
	}
	if req.TargetID == 0 {
		return nil, errors.New("target_id wajib diisi")
	}
	if req.ReportType == "" {
		return nil, errors.New("report_type wajib diisi")
	}
	if !validReportTypes[req.ReportType] {
		return nil, errors.New("report_type tidak valid: harus harassment, violence, hate_speech, spam, inappropriate, misinformation, copyright, atau other")
	}
	if req.ReportType == "other" && (req.Description == nil || *req.Description == "") {
		return nil, errors.New("deskripsi wajib diisi ketika jenis laporan adalah 'other'")
	}

	// Prevent duplicate pending reports on the same content by the same user
	existing, err := s.reportRepo.FindExistingActiveReport(reporterID, req.TargetType, req.TargetID)
	if err == nil && existing != nil {
		return nil, errors.New("Anda sudah melaporkan konten ini dan laporan masih dalam proses")
	}

	report := &models.Report{
		ReporterID:  reporterID,
		TargetType:  req.TargetType,
		TargetID:    req.TargetID,
		ReportType:  req.ReportType,
		Description: req.Description,
	}

	if err := s.reportRepo.CreateReport(report); err != nil {
		return nil, errors.New("gagal mengirimkan laporan")
	}
	return report, nil
}

func (s *reportService) GetMyReports(reporterID uint, page, limit int) ([]models.Report, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	reports, total, err := s.reportRepo.FindMyReports(reporterID, page, limit)
	if err != nil {
		return nil, 0, errors.New("gagal mengambil daftar laporan")
	}
	return reports, total, nil
}
