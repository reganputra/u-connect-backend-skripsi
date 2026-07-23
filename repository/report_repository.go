package repository

import (
	"github.com/reganputra/skripsi-backend/constant"
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type ReportRepository interface {
	CreateReport(r *models.Report) error
	FindReports(page, limit int, status string) ([]models.Report, int64, error)
	FindReportByID(id uint) (*models.Report, error)
	FindMyReports(reporterID uint, page, limit int) ([]models.Report, int64, error)
	FindExistingActiveReport(reporterID uint, targetType string, targetID uint) (*models.Report, error)
	UpdateReport(r *models.Report) error
}

type reportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) CreateReport(report *models.Report) error {
	return r.db.Create(report).Error
}

func (r *reportRepository) FindReports(page, limit int, status string) ([]models.Report, int64, error) {
	var reports []models.Report
	var total int64
	offset := (page - 1) * limit

	query := r.db.Model(&models.Report{}).Preload("Reporter")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&reports).Error
	return reports, total, err
}

func (r *reportRepository) FindReportByID(id uint) (*models.Report, error) {
	var report models.Report
	err := r.db.
		Preload("Reporter").
		Preload("ResolvedBy").
		First(&report, id).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *reportRepository) FindMyReports(reporterID uint, page, limit int) ([]models.Report, int64, error) {
	var reports []models.Report
	var total int64
	offset := (page - 1) * limit

	if err := r.db.Model(&models.Report{}).Where("reporter_id = ?", reporterID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.
		Where("reporter_id = ?", reporterID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&reports).Error
	return reports, total, err
}

func (r *reportRepository) FindExistingActiveReport(reporterID uint, targetType string, targetID uint) (*models.Report, error) {
	var report models.Report
	err := r.db.
		Where("reporter_id = ? AND target_type = ? AND target_id = ? AND status = ?", reporterID, targetType, targetID, constant.StatusPending).
		First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *reportRepository) UpdateReport(report *models.Report) error {
	return r.db.Save(report).Error
}
