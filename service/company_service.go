package service

import (
	"errors"
	"strings"

	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
)

// ─── Interface ────────────────────────────────────────────────────────────────

type CompanyService interface {
	CreateOrJoinCompanyProfile(userID uint, req CompanyProfileRequest) (*models.CompanyProfile, bool, error)
	GetCompanyProfile(userID uint) (*models.CompanyProfile, error)
	UpdateCompanyProfile(userID uint, req CompanyProfileRequest) (*models.CompanyProfile, error)
	ChangeCompanyAffiliation(userID uint, companyName string) (*models.CompanyProfile, bool, error)
	DeleteCompanyProfile(userID uint) error
}

// ─── DTO ──────────────────────────────────────────────────────────────────────

type CompanyProfileRequest struct {
	CompanyName  *string `json:"company_name"`
	Description  *string `json:"description"`
	IndustryType *string `json:"industry_type"`
	Location     *string `json:"location"`
	EmployeeSize *int    `json:"employee_size"`
	WebsiteURL   *string `json:"website_url"`
}

// ─── Implementation ───────────────────────────────────────────────────────────

type companyService struct {
	companyRepo repository.CompanyRepository
	userRepo    repository.UserRepository
	profileRepo repository.ProfileRepository
}

func NewCompanyService(companyRepo repository.CompanyRepository, userRepo repository.UserRepository, profileRepo repository.ProfileRepository) CompanyService {
	return &companyService{
		companyRepo: companyRepo,
		userRepo:    userRepo,
		profileRepo: profileRepo,
	}
}

func (s *companyService) ensureDirectoryProfile(userID uint) error {
	if s.profileRepo == nil {
		return nil
	}
	if err := s.profileRepo.EnsureProfileExists(userID); err != nil {
		return errors.New("gagal menyiapkan profil pengguna")
	}
	return nil
}

func (s *companyService) CreateOrJoinCompanyProfile(userID uint, req CompanyProfileRequest) (*models.CompanyProfile, bool, error) {
	user, err := s.userRepo.FindUserByID(userID)
	if err != nil {
		return nil, false, errors.New("pengguna tidak ditemukan")
	}

	if req.EmployeeSize != nil && *req.EmployeeSize < 0 {
		return nil, false, errors.New("jumlah karyawan harus nol atau positif")
	}

	var companyName string
	if user.CompanyName != nil {
		companyName = strings.TrimSpace(*user.CompanyName)
	}
	if companyName == "" {
		if req.CompanyName == nil || strings.TrimSpace(*req.CompanyName) == "" {
			return nil, false, errors.New("nama perusahaan wajib diisi saat pertama kali membuat atau bergabung profil perusahaan")
		}
		companyName = strings.TrimSpace(*req.CompanyName)
	}

	existing, err := s.companyRepo.FindCompanyProfileByName(companyName)
	if err == nil && existing != nil {
		if user.CompanyName == nil || strings.TrimSpace(*user.CompanyName) == "" {
			if err := s.userRepo.UpdateUserCompanyName(userID, companyName); err != nil {
				return nil, false, errors.New("gagal menyimpan nama perusahaan pada akun")
			}
		}
		if err := s.ensureDirectoryProfile(userID); err != nil {
			return nil, false, err
		}
		return existing, false, nil
	}

	profile := &models.CompanyProfile{
		CompanyName:  companyName,
		Description:  req.Description,
		IndustryType: req.IndustryType,
		Location:     req.Location,
		EmployeeSize: req.EmployeeSize,
		WebsiteURL:   req.WebsiteURL,
	}

	if err := s.companyRepo.CreateCompanyProfile(profile); err != nil {
		return nil, false, errors.New("gagal membuat profil perusahaan")
	}
	if user.CompanyName == nil || strings.TrimSpace(*user.CompanyName) == "" {
		if err := s.userRepo.UpdateUserCompanyName(userID, companyName); err != nil {
			return nil, false, errors.New("gagal menyimpan nama perusahaan pada akun")
		}
	}
	if err := s.ensureDirectoryProfile(userID); err != nil {
		return nil, false, err
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
	if req.Description != nil {
		profile.Description = req.Description
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

func (s *companyService) ChangeCompanyAffiliation(userID uint, companyName string) (*models.CompanyProfile, bool, error) {
	newName := strings.TrimSpace(companyName)
	if newName == "" {
		return nil, false, errors.New("nama perusahaan wajib diisi")
	}

	user, err := s.userRepo.FindUserByID(userID)
	if err != nil {
		return nil, false, errors.New("pengguna tidak ditemukan")
	}

	currentName := ""
	if user.CompanyName != nil {
		currentName = strings.TrimSpace(*user.CompanyName)
	}

	if currentName == newName {
		profile, err := s.companyRepo.FindCompanyProfileByName(newName)
		if err != nil {
			return nil, false, errors.New("profil perusahaan tidak ditemukan")
		}
		if err := s.ensureDirectoryProfile(userID); err != nil {
			return nil, false, err
		}
		return profile, false, nil
	}

	if existing, err := s.companyRepo.FindCompanyProfileByName(newName); err == nil && existing != nil {
		if err := s.userRepo.UpdateUserCompanyName(userID, newName); err != nil {
			return nil, false, errors.New("gagal memperbarui afiliasi perusahaan")
		}
		if err := s.ensureDirectoryProfile(userID); err != nil {
			return nil, false, err
		}
		return existing, true, nil
	}

	profile := &models.CompanyProfile{CompanyName: newName}
	if err := s.companyRepo.CreateCompanyProfile(profile); err != nil {
		return nil, false, errors.New("gagal membuat profil perusahaan")
	}
	if err := s.userRepo.UpdateUserCompanyName(userID, newName); err != nil {
		return nil, false, errors.New("gagal memperbarui afiliasi perusahaan")
	}
	if err := s.ensureDirectoryProfile(userID); err != nil {
		return nil, false, err
	}

	return profile, false, nil
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
