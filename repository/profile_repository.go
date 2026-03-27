package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type ProfileRepository interface {
	CreateProfile(profile *models.UserProfile) error
	FindProfileByUserID(userID uint) (*models.UserProfile, error)
	UpdateProfile(profile *models.UserProfile) error
	DeleteProfileByUserID(userID uint) error
	AddExperience(exp *models.UserExperience) error
	FindExperienceByID(id uint) (*models.UserExperience, error)
	UpdateExperience(exp *models.UserExperience) error
	DeleteExperience(id uint) error
}

type profileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) ProfileRepository {
	return &profileRepository{db: db}
}

func (r *profileRepository) CreateProfile(profile *models.UserProfile) error {
	return r.db.Create(profile).Error
}

func (r *profileRepository) FindProfileByUserID(userID uint) (*models.UserProfile, error) {
	var profile models.UserProfile
	result := r.db.
		Preload("User").
		Preload("Experiences").
		Where("user_id = ?", userID).
		First(&profile)
	if result.Error != nil {
		return nil, result.Error
	}
	return &profile, nil
}

func (r *profileRepository) UpdateProfile(profile *models.UserProfile) error {
	return r.db.Save(profile).Error
}

func (r *profileRepository) DeleteProfileByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.UserProfile{}).Error
}

func (r *profileRepository) AddExperience(exp *models.UserExperience) error {
	return r.db.Create(exp).Error
}

func (r *profileRepository) FindExperienceByID(id uint) (*models.UserExperience, error) {
	var exp models.UserExperience
	if err := r.db.First(&exp, id).Error; err != nil {
		return nil, err
	}
	return &exp, nil
}

func (r *profileRepository) UpdateExperience(exp *models.UserExperience) error {
	return r.db.Save(exp).Error
}

func (r *profileRepository) DeleteExperience(id uint) error {
	return r.db.Delete(&models.UserExperience{}, id).Error
}
