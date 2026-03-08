package repository

import (
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/models"
)

func CreateJobApplication(app *models.JobApplication) error {
	return config.DB.Create(app).Error
}

func FindJobApplication(jobID, userID uint) (*models.JobApplication, error) {
	var app models.JobApplication
	err := config.DB.Where("job_id = ? AND user_id = ?", jobID, userID).First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func FindApplicantsByJobID(jobID uint) ([]models.JobApplication, error) {
	var apps []models.JobApplication
	err := config.DB.Preload("User").Where("job_id = ?", jobID).Order("created_at DESC").Find(&apps).Error
	return apps, err
}

func FindApplicationsByUserID(userID uint) ([]models.JobApplication, error) {
	var apps []models.JobApplication
	err := config.DB.Preload("Job").Preload("Job.User").Where("user_id = ?", userID).Order("created_at DESC").Find(&apps).Error
	return apps, err
}

func FindJobApplicationByID(id uint) (*models.JobApplication, error) {
	var app models.JobApplication
	err := config.DB.Preload("Job").First(&app, id).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func UpdateJobApplication(app *models.JobApplication) error {
	return config.DB.Save(app).Error
}
