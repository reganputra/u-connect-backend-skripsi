package repository

import (
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/models"
)

func CreateProfile(profile *models.UserProfile) error {
	return config.DB.Create(profile).Error
}

func FindProfileByUserID(userID uint) (*models.UserProfile, error) {
	var profile models.UserProfile
	result := config.DB.
		Preload("User").
		Preload("Experiences").
		Where("user_id = ?", userID).
		First(&profile)
	if result.Error != nil {
		return nil, result.Error
	}
	return &profile, nil
}

func UpdateProfile(profile *models.UserProfile) error {
	return config.DB.Save(profile).Error
}

func DeleteProfileByUserID(userID uint) error {
	return config.DB.Where("user_id = ?", userID).Delete(&models.UserProfile{}).Error
}

// Experience operations
func AddExperience(exp *models.UserExperience) error {
	return config.DB.Create(exp).Error
}

func FindExperienceByID(id uint) (*models.UserExperience, error) {
	var exp models.UserExperience
	if err := config.DB.First(&exp, id).Error; err != nil {
		return nil, err
	}
	return &exp, nil
}

func UpdateExperience(exp *models.UserExperience) error {
	return config.DB.Save(exp).Error
}

func DeleteExperience(id uint) error {
	return config.DB.Delete(&models.UserExperience{}, id).Error
}
