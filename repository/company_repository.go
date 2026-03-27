package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type CompanyRepository interface {
	CreateCompanyProfile(profile *models.CompanyProfile) error
	FindCompanyProfileByName(name string) (*models.CompanyProfile, error)
	FindCompanyProfileByID(id uint) (*models.CompanyProfile, error)
	UpdateCompanyProfile(profile *models.CompanyProfile) error
	DeleteCompanyProfile(id uint) error
}

type companyRepository struct {
	db *gorm.DB
}

func NewCompanyRepository(db *gorm.DB) CompanyRepository {
	return &companyRepository{db: db}
}

func (r *companyRepository) CreateCompanyProfile(profile *models.CompanyProfile) error {
	return r.db.Create(profile).Error
}

func (r *companyRepository) FindCompanyProfileByName(name string) (*models.CompanyProfile, error) {
	var profile models.CompanyProfile
	result := r.db.Where("company_name = ?", name).First(&profile)
	if result.Error != nil {
		return nil, result.Error
	}
	return &profile, nil
}

func (r *companyRepository) FindCompanyProfileByID(id uint) (*models.CompanyProfile, error) {
	var profile models.CompanyProfile
	if err := r.db.First(&profile, id).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *companyRepository) UpdateCompanyProfile(profile *models.CompanyProfile) error {
	return r.db.Save(profile).Error
}

func (r *companyRepository) DeleteCompanyProfile(id uint) error {
	return r.db.Delete(&models.CompanyProfile{}, id).Error
}
