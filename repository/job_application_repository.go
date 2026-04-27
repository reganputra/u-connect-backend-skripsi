package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type JobApplicationRepository interface {
	CreateJobApplication(app *models.JobApplication) error
	FindJobApplication(jobID, userID uint) (*models.JobApplication, error)
	FindApplicantsByJobID(jobID uint) ([]models.JobApplication, error)
	FindApplicationsByUserID(userID uint) ([]models.JobApplication, error)
	CountApplicationsByUserID(userID uint) (int64, error)
	FindJobApplicationByID(id uint) (*models.JobApplication, error)
	UpdateJobApplication(app *models.JobApplication) error
}

type jobApplicationRepository struct {
	db *gorm.DB
}

func NewJobApplicationRepository(db *gorm.DB) JobApplicationRepository {
	return &jobApplicationRepository{db: db}
}

func (r *jobApplicationRepository) CreateJobApplication(app *models.JobApplication) error {
	return r.db.Create(app).Error
}

func (r *jobApplicationRepository) FindJobApplication(jobID, userID uint) (*models.JobApplication, error) {
	var app models.JobApplication
	err := r.db.Where("job_id = ? AND user_id = ?", jobID, userID).First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *jobApplicationRepository) FindApplicantsByJobID(jobID uint) ([]models.JobApplication, error) {
	var apps []models.JobApplication
	err := r.db.Preload("User").Preload("User.Profile").Where("job_id = ? AND status <> ?", jobID, "withdrawn").Order("created_at DESC").Find(&apps).Error
	return apps, err
}

func (r *jobApplicationRepository) FindApplicationsByUserID(userID uint) ([]models.JobApplication, error) {
	var apps []models.JobApplication
	err := r.db.Preload("Job").Preload("Job.User").Where("user_id = ?", userID).Order("created_at DESC").Find(&apps).Error
	return apps, err
}

func (r *jobApplicationRepository) CountApplicationsByUserID(userID uint) (int64, error) {
	var total int64
	err := r.db.Model(&models.JobApplication{}).Where("user_id = ?", userID).Count(&total).Error
	return total, err
}

func (r *jobApplicationRepository) FindJobApplicationByID(id uint) (*models.JobApplication, error) {
	var app models.JobApplication
	err := r.db.Preload("Job").Preload("Job.Company").First(&app, id).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *jobApplicationRepository) UpdateJobApplication(app *models.JobApplication) error {
	return r.db.Save(app).Error
}
