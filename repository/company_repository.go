package repository

import (
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/models"
)

func CreateCompanyProfile(profile *models.CompanyProfile) error {
	return config.DB.Create(profile).Error
}

func FindCompanyProfileByName(name string) (*models.CompanyProfile, error) {
	var profile models.CompanyProfile
	result := config.DB.Where("company_name = ?", name).First(&profile)
	if result.Error != nil {
		return nil, result.Error
	}
	return &profile, nil
}

func FindCompanyProfileByID(id uint) (*models.CompanyProfile, error) {
	var profile models.CompanyProfile
	if err := config.DB.First(&profile, id).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func UpdateCompanyProfile(profile *models.CompanyProfile) error {
	return config.DB.Save(profile).Error
}

func DeleteCompanyProfile(id uint) error {
	return config.DB.Delete(&models.CompanyProfile{}, id).Error
}
