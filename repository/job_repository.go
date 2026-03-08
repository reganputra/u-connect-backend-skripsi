package repository

import (
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

func CreateJob(job *models.Job) error {
	return config.DB.Create(job).Error
}

func FindJobs(search, jobType, status string, offset, limit int) ([]models.Job, int64, error) {
	var jobs []models.Job
	var total int64

	query := config.DB.Model(&models.Job{}).Preload("User")

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

	query.Count(&total)
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&jobs).Error
	return jobs, total, err
}

func FindJobByID(id uint) (*models.Job, error) {
	var job models.Job
	err := config.DB.Preload("User").First(&job, id).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func UpdateJob(job *models.Job) error {
	return config.DB.Save(job).Error
}

func DeleteJob(id uint) error {
	return config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("job_id = ?", id).Delete(&models.JobApplication{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Job{}, id).Error
	})
}
