package service

import (
	"errors"

	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
)

// ─── Interface ────────────────────────────────────────────────────────────────

type CompanyService interface {
	CreateOrJoinCompanyProfile(userID uint, req CompanyProfileRequest) (*models.CompanyProfile, bool, error)
	GetCompanyProfile(userID uint) (*models.CompanyProfile, error)
	UpdateCompanyProfile(userID uint, req CompanyProfileRequest) (*models.CompanyProfile, error)
	DeleteCompanyProfile(userID uint) error
}

// ─── DTO ──────────────────────────────────────────────────────────────────────

type CompanyProfileRequest struct {
	IndustryType *string `json:"industry_type"`
	Location     *string `json:"location"`
	EmployeeSize *int    `json:"employee_size"`
	WebsiteURL   *string `json:"website_url"`
}

// ─── Implementation ───────────────────────────────────────────────────────────

type companyService struct {
	companyRepo repository.CompanyRepository
	userRepo    repository.UserRepository
}

func NewCompanyService(companyRepo repository.CompanyRepository, userRepo repository.UserRepository) CompanyService {
	return &companyService{
		companyRepo: companyRepo,
		userRepo:    userRepo,
	}
}

func (s *companyService) CreateOrJoinCompanyProfile(userID uint, req CompanyProfileRequest) (*models.CompanyProfile, bool, error) {
	user, err := s.userRepo.FindUserByID(userID)
	if err != nil {
		return nil, false, errors.New("pengguna tidak ditemukan")
	}
	if user.CompanyName == nil || *user.CompanyName == "" {
		return nil, false, errors.New("nama perusahaan belum diatur pada akun Anda")
	}

	if req.EmployeeSize != nil && *req.EmployeeSize < 0 {
		return nil, false, errors.New("jumlah karyawan harus nol atau positif")
	}

	companyName := *user.CompanyName

	existing, err := s.companyRepo.FindCompanyProfileByName(companyName)
	if err == nil && existing != nil {
		return existing, false, nil
	}

	profile := &models.CompanyProfile{
		CompanyName:  companyName,
		IndustryType: req.IndustryType,
		Location:     req.Location,
		EmployeeSize: req.EmployeeSize,
		WebsiteURL:   req.WebsiteURL,
	}

	if err := s.companyRepo.CreateCompanyProfile(profile); err != nil {
		return nil, false, errors.New("gagal membuat profil perusahaan")
	}
	return profile, true, nil
}

func (s *companyService) GetCompanyProfile(userID uint) (*models.CompanyProfile, error) {
	user, err := s.userRepo.FindUserByID(userID)
	if err != nil {
		return nil, errors.New("pengguna tidak ditemukan")
	}
	if user.CompanyName == nil || *user.CompanyName == "" {
		return nil, errors.New("nama perusahaan belum diatur pada akun Anda")
	}

	profile, err := s.companyRepo.FindCompanyProfileByName(*user.CompanyName)
	if err != nil {
		return nil, errors.New("profil perusahaan tidak ditemukan")
	}
	return profile, nil
}

func (s *companyService) UpdateCompanyProfile(userID uint, req CompanyProfileRequest) (*models.CompanyProfile, error) {
	user, err := s.userRepo.FindUserByID(userID)
	if err != nil {
		return nil, errors.New("pengguna tidak ditemukan")
	}
	if user.CompanyName == nil || *user.CompanyName == "" {
		return nil, errors.New("nama perusahaan belum diatur pada akun Anda")
	}

	profile, err := s.companyRepo.FindCompanyProfileByName(*user.CompanyName)
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

	if err := s.companyRepo.UpdateCompanyProfile(profile); err != nil {
		return nil, errors.New("gagal memperbarui profil perusahaan")
	}
	return profile, nil
}

func (s *companyService) DeleteCompanyProfile(userID uint) error {
	user, err := s.userRepo.FindUserByID(userID)
	if err != nil {
		return errors.New("pengguna tidak ditemukan")
	}
	if user.CompanyName == nil || *user.CompanyName == "" {
		return errors.New("nama perusahaan belum diatur pada akun Anda")
	}

	profile, err := s.companyRepo.FindCompanyProfileByName(*user.CompanyName)
	if err != nil {
		return errors.New("profil perusahaan tidak ditemukan")
	}

	return s.companyRepo.DeleteCompanyProfile(profile.ID)
}
