package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type JobRepository interface {
	CreateJob(job *models.Job) error
	FindJobs(search, jobType, status string, offset, limit int) ([]models.Job, int64, error)
	FindJobsByOwner(userID uint, offset, limit int) ([]models.Job, int64, error)
	FindJobByID(id uint) (*models.Job, error)
	UpdateJob(job *models.Job) error
	DeleteJob(id uint) error
}

type jobRepository struct {
	db *gorm.DB
}

func NewJobRepository(db *gorm.DB) JobRepository {
	return &jobRepository{db: db}
}

func (r *jobRepository) CreateJob(job *models.Job) error {
	return r.db.Create(job).Error
}

func (r *jobRepository) FindJobs(search, jobType, status string, offset, limit int) ([]models.Job, int64, error) {
	var jobs []models.Job
	var total int64

	query := r.db.Model(&models.Job{}).Preload("User").Preload("Company")

	if search != "" {
		like := "%" + search + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ? OR company_name ILIKE ?", like, like, like)
	}
	if jobType != "" {
		query = query.Where("job_type = ?", jobType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&jobs).Error
	return jobs, total, err
}

func (r *jobRepository) FindJobsByOwner(userID uint, offset, limit int) ([]models.Job, int64, error) {
	var jobs []models.Job
	var total int64

	query := r.db.Model(&models.Job{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.
		Preload("User").
		Preload("Company").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&jobs).Error

	return jobs, total, err
}

func (r *jobRepository) FindJobByID(id uint) (*models.Job, error) {
	var job models.Job
	err := r.db.Preload("User").Preload("Company").First(&job, id).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *jobRepository) UpdateJob(job *models.Job) error {
	return r.db.Save(job).Error
}

func (r *jobRepository) DeleteJob(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("job_id = ?", id).Delete(&models.JobApplication{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Job{}, id).Error
	})
}
