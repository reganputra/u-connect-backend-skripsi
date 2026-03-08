package service

import (
	"errors"

	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
)

type CompanyProfileRequest struct {
	IndustryType *string `json:"industry_type"`
	Location     *string `json:"location"`
	EmployeeSize *int    `json:"employee_size"`
	WebsiteURL   *string `json:"website_url"`
}

func CreateOrJoinCompanyProfile(userID uint, req CompanyProfileRequest) (*models.CompanyProfile, bool, error) {
	// Load user to get company_name from registration
	user, err := repository.FindUserByID(userID)
	if err != nil {
		return nil, false, errors.New("pengguna tidak ditemukan")
	}
	if user.CompanyName == nil || *user.CompanyName == "" {
		return nil, false, errors.New("nama perusahaan belum diatur pada akun Anda")
	}

	// Validate employee_size if provided
	if req.EmployeeSize != nil && *req.EmployeeSize < 0 {
		return nil, false, errors.New("jumlah karyawan harus nol atau positif")
	}

	companyName := *user.CompanyName

	// Check if profile already exists for this company
	existing, err := repository.FindCompanyProfileByName(companyName)
	if err == nil && existing != nil {
		// Profile already exists — return it (user auto-joins)
		return existing, false, nil
	}

	// Create new profile
	profile := &models.CompanyProfile{
		CompanyName:  companyName,
		IndustryType: req.IndustryType,
		Location:     req.Location,
		EmployeeSize: req.EmployeeSize,
		WebsiteURL:   req.WebsiteURL,
	}

	if err := repository.CreateCompanyProfile(profile); err != nil {
		return nil, false, errors.New("gagal membuat profil perusahaan")
	}
	return profile, true, nil
}

func GetCompanyProfile(userID uint) (*models.CompanyProfile, error) {
	user, err := repository.FindUserByID(userID)
	if err != nil {
		return nil, errors.New("pengguna tidak ditemukan")
	}
	if user.CompanyName == nil || *user.CompanyName == "" {
		return nil, errors.New("nama perusahaan belum diatur pada akun Anda")
	}

	profile, err := repository.FindCompanyProfileByName(*user.CompanyName)
	if err != nil {
		return nil, errors.New("profil perusahaan tidak ditemukan")
	}
	return profile, nil
}

func UpdateCompanyProfile(userID uint, req CompanyProfileRequest) (*models.CompanyProfile, error) {
	user, err := repository.FindUserByID(userID)
	if err != nil {
		return nil, errors.New("pengguna tidak ditemukan")
	}
	if user.CompanyName == nil || *user.CompanyName == "" {
		return nil, errors.New("nama perusahaan belum diatur pada akun Anda")
	}

	profile, err := repository.FindCompanyProfileByName(*user.CompanyName)
	if err != nil {
		return nil, errors.New("profil perusahaan tidak ditemukan")
	}

	if req.EmployeeSize != nil {
		if *req.EmployeeSize < 0 {
			return nil, errors.New("jumlah karyawan harus nol atau positif")
		}
		profile.EmployeeSize = req.EmployeeSize
	}
	if req.IndustryType != nil {
		profile.IndustryType = req.IndustryType
	}
	if req.Location != nil {
		profile.Location = req.Location
	}
	if req.WebsiteURL != nil {
		profile.WebsiteURL = req.WebsiteURL
	}

	if err := repository.UpdateCompanyProfile(profile); err != nil {
		return nil, errors.New("gagal memperbarui profil perusahaan")
	}
	return profile, nil
}

func DeleteCompanyProfile(userID uint) error {
	user, err := repository.FindUserByID(userID)
	if err != nil {
		return errors.New("pengguna tidak ditemukan")
	}
	if user.CompanyName == nil || *user.CompanyName == "" {
		return errors.New("nama perusahaan belum diatur pada akun Anda")
	}

	profile, err := repository.FindCompanyProfileByName(*user.CompanyName)
	if err != nil {
		return errors.New("profil perusahaan tidak ditemukan")
	}

	return repository.DeleteCompanyProfile(profile.ID)
}
