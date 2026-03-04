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
		return nil, false, errors.New("user not found")
	}
	if user.CompanyName == nil || *user.CompanyName == "" {
		return nil, false, errors.New("company_name is not set on your account")
	}

	// Validate employee_size if provided
	if req.EmployeeSize != nil && *req.EmployeeSize < 0 {
		return nil, false, errors.New("employee_size must be zero or positive")
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
		return nil, false, errors.New("failed to create company profile")
	}
	return profile, true, nil
}

func GetCompanyProfile(userID uint) (*models.CompanyProfile, error) {
	user, err := repository.FindUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if user.CompanyName == nil || *user.CompanyName == "" {
		return nil, errors.New("company_name is not set on your account")
	}

	profile, err := repository.FindCompanyProfileByName(*user.CompanyName)
	if err != nil {
		return nil, errors.New("company profile not found")
	}
	return profile, nil
}

func UpdateCompanyProfile(userID uint, req CompanyProfileRequest) (*models.CompanyProfile, error) {
	user, err := repository.FindUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if user.CompanyName == nil || *user.CompanyName == "" {
		return nil, errors.New("company_name is not set on your account")
	}

	profile, err := repository.FindCompanyProfileByName(*user.CompanyName)
	if err != nil {
		return nil, errors.New("company profile not found")
	}

	if req.EmployeeSize != nil {
		if *req.EmployeeSize < 0 {
			return nil, errors.New("employee_size must be zero or positive")
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
		return nil, errors.New("failed to update company profile")
	}
	return profile, nil
}

func DeleteCompanyProfile(userID uint) error {
	user, err := repository.FindUserByID(userID)
	if err != nil {
		return errors.New("user not found")
	}
	if user.CompanyName == nil || *user.CompanyName == "" {
		return errors.New("company_name is not set on your account")
	}

	profile, err := repository.FindCompanyProfileByName(*user.CompanyName)
	if err != nil {
		return errors.New("company profile not found")
	}

	return repository.DeleteCompanyProfile(profile.ID)
}
