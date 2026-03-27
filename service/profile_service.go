package service

import (
	"errors"

	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
)

var validProfileJobStatuses = map[string]bool{
	"employed":         true,
	"entrepreneur":     true,
	"continuing_study": true,
	"unemployed":       true,
	"freelance":        true,
	"student":          true,
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type ProfileRequest struct {
	ProfilePicture *string `json:"profile_picture"`
	Bio            *string `json:"bio"`
	Location       *string `json:"location"`

	// Professional
	JobStatus    *string `json:"job_status"`
	Position     *string `json:"position"`
	CompanyName  *string `json:"company_name"`
	IndustryName *string `json:"industry_name"`
	IndustryType *string `json:"industry_type"`
	YearFounding *int    `json:"year_founding"`
	Salary       *int    `json:"salary"`

	// Academic
	EducationalLevel       *string `json:"educational_level"`
	AdvancedStudyProgram   *string `json:"advanced_study_program"`
	InstitutionName        *string `json:"institution_name"`
	ExpectedGraduationYear *int    `json:"expected_graduation_year"`

	// Skills & Interests
	Skills    *string `json:"skills"`
	Interests *string `json:"interests"`

	// Mentorship (alumni only)
	MentorQuota       *int    `json:"mentor_quota"`
	MentorDescription *string `json:"mentor_description"`

	// Additional
	StatusDescription *string `json:"status_description"`
}

type ExperienceRequest struct {
	CompanyName string  `json:"company_name"`
	Position    string  `json:"position"`
	StartYear   int     `json:"start_year"`
	EndYear     *int    `json:"end_year"`
	Description *string `json:"description"`
}

// ─── Interface ────────────────────────────────────────────────────────────────

type ProfileService interface {
	CreateProfile(userID uint, req ProfileRequest) (*models.UserProfile, error)
	GetProfile(userID uint) (*models.UserProfile, error)
	UpdateProfile(userID uint, req ProfileRequest) (*models.UserProfile, error)
	DeleteProfile(userID uint) error
	AddExperience(userID uint, req ExperienceRequest) (*models.UserExperience, error)
	UpdateExperience(userID uint, expID uint, req ExperienceRequest) (*models.UserExperience, error)
	DeleteExperience(userID uint, expID uint) error
	UpdateProfilePicture(userID uint, pictureURL string) error
}

// ─── Implementation ───────────────────────────────────────────────────────────

type profileService struct {
	profileRepo repository.ProfileRepository
}

func NewProfileService(profileRepo repository.ProfileRepository) ProfileService {
	return &profileService{profileRepo: profileRepo}
}

// ─── Validation helpers ───────────────────────────────────────────────────────

func validateJobStatus(req ProfileRequest) error {
	if req.JobStatus == nil {
		return nil
	}
	status := *req.JobStatus
	if !validProfileJobStatuses[status] {
		return errors.New("status pekerjaan tidak valid: harus employed, entrepreneur, continuing_study, unemployed, freelance, atau student")
	}
	switch status {
	case "employed":
		if req.Position == nil || *req.Position == "" {
			return errors.New("posisi wajib diisi ketika status pekerjaan adalah employed")
		}
		if req.CompanyName == nil || *req.CompanyName == "" {
			return errors.New("nama perusahaan wajib diisi ketika status pekerjaan adalah employed")
		}
	case "entrepreneur":
		if req.IndustryName == nil || *req.IndustryName == "" {
			return errors.New("nama industri wajib diisi ketika status pekerjaan adalah entrepreneur")
		}
	case "continuing_study":
		if req.EducationalLevel == nil || *req.EducationalLevel == "" {
			return errors.New("jenjang pendidikan wajib diisi ketika status pekerjaan adalah continuing_study")
		}
		if req.AdvancedStudyProgram == nil || *req.AdvancedStudyProgram == "" {
			return errors.New("program studi lanjut wajib diisi ketika status pekerjaan adalah continuing_study")
		}
		if req.InstitutionName == nil || *req.InstitutionName == "" {
			return errors.New("nama institusi wajib diisi ketika status pekerjaan adalah continuing_study")
		}
	case "unemployed", "freelance":
		if req.StatusDescription == nil || *req.StatusDescription == "" {
			return errors.New("deskripsi status wajib diisi ketika status pekerjaan adalah unemployed atau freelance")
		}
	}
	return nil
}

func nullFieldsByStatus(profile *models.UserProfile) {
	if profile.JobStatus == nil {
		return
	}
	status := *profile.JobStatus
	if status != "employed" {
		profile.Position = nil
		if status != "entrepreneur" {
			profile.CompanyName = nil
		}
	}
	if status != "entrepreneur" {
		profile.IndustryName = nil
		profile.IndustryType = nil
		profile.YearFounding = nil
	}
	if status != "continuing_study" {
		profile.EducationalLevel = nil
		profile.AdvancedStudyProgram = nil
		profile.InstitutionName = nil
		profile.ExpectedGraduationYear = nil
	}
	if status != "unemployed" && status != "freelance" {
		profile.StatusDescription = nil
	}
}

// ─── Service methods ──────────────────────────────────────────────────────────

func (s *profileService) CreateProfile(userID uint, req ProfileRequest) (*models.UserProfile, error) {
	existing, _ := s.profileRepo.FindProfileByUserID(userID)
	if existing != nil {
		return nil, errors.New("profil sudah ada")
	}

	if err := validateJobStatus(req); err != nil {
		return nil, err
	}

	profile := &models.UserProfile{
		UserID:                 userID,
		Bio:                    req.Bio,
		Location:               req.Location,
		JobStatus:              req.JobStatus,
		Position:               req.Position,
		CompanyName:            req.CompanyName,
		IndustryName:           req.IndustryName,
		IndustryType:           req.IndustryType,
		YearFounding:           req.YearFounding,
		Salary:                 req.Salary,
		EducationalLevel:       req.EducationalLevel,
		AdvancedStudyProgram:   req.AdvancedStudyProgram,
		InstitutionName:        req.InstitutionName,
		ExpectedGraduationYear: req.ExpectedGraduationYear,
		Skills:                 req.Skills,
		Interests:              req.Interests,
		MentorQuota:            req.MentorQuota,
		MentorDescription:      req.MentorDescription,
		StatusDescription:      req.StatusDescription,
	}

	if req.ProfilePicture != nil {
		profile.ProfilePicture = *req.ProfilePicture
	}

	nullFieldsByStatus(profile)

	if err := s.profileRepo.CreateProfile(profile); err != nil {
		return nil, errors.New("gagal membuat profil")
	}
	return profile, nil
}

func (s *profileService) GetProfile(userID uint) (*models.UserProfile, error) {
	profile, err := s.profileRepo.FindProfileByUserID(userID)
	if err != nil {
		return nil, errors.New("profil tidak ditemukan")
	}
	return profile, nil
}

func (s *profileService) UpdateProfile(userID uint, req ProfileRequest) (*models.UserProfile, error) {
	profile, err := s.profileRepo.FindProfileByUserID(userID)
	if err != nil {
		return nil, errors.New("profil tidak ditemukan")
	}

	if req.Bio != nil {
		profile.Bio = req.Bio
	}
	if req.Location != nil {
		profile.Location = req.Location
	}
	if req.JobStatus != nil {
		profile.JobStatus = req.JobStatus
	}
	if req.Position != nil {
		profile.Position = req.Position
	}
	if req.CompanyName != nil {
		profile.CompanyName = req.CompanyName
	}
	if req.IndustryName != nil {
		profile.IndustryName = req.IndustryName
	}
	if req.IndustryType != nil {
		profile.IndustryType = req.IndustryType
	}
	if req.YearFounding != nil {
		profile.YearFounding = req.YearFounding
	}
	if req.Salary != nil {
		profile.Salary = req.Salary
	}
	if req.EducationalLevel != nil {
		profile.EducationalLevel = req.EducationalLevel
	}
	if req.AdvancedStudyProgram != nil {
		profile.AdvancedStudyProgram = req.AdvancedStudyProgram
	}
	if req.InstitutionName != nil {
		profile.InstitutionName = req.InstitutionName
	}
	if req.ExpectedGraduationYear != nil {
		profile.ExpectedGraduationYear = req.ExpectedGraduationYear
	}
	if req.Skills != nil {
		profile.Skills = req.Skills
	}
	if req.Interests != nil {
		profile.Interests = req.Interests
	}
	if req.MentorQuota != nil {
		profile.MentorQuota = req.MentorQuota
	}
	if req.MentorDescription != nil {
		profile.MentorDescription = req.MentorDescription
	}
	if req.ProfilePicture != nil {
		profile.ProfilePicture = *req.ProfilePicture
	}
	if req.StatusDescription != nil {
		profile.StatusDescription = req.StatusDescription
	}

	if err := validateJobStatus(req); err != nil {
		return nil, err
	}
	nullFieldsByStatus(profile)

	if err := s.profileRepo.UpdateProfile(profile); err != nil {
		return nil, errors.New("gagal memperbarui profil")
	}
	return profile, nil
}

func (s *profileService) DeleteProfile(userID uint) error {
	_, err := s.profileRepo.FindProfileByUserID(userID)
	if err != nil {
		return errors.New("profil tidak ditemukan")
	}
	return s.profileRepo.DeleteProfileByUserID(userID)
}

func (s *profileService) AddExperience(userID uint, req ExperienceRequest) (*models.UserExperience, error) {
	if req.CompanyName == "" || req.Position == "" || req.StartYear == 0 {
		return nil, errors.New("nama perusahaan, posisi, dan tahun mulai wajib diisi")
	}

	profile, err := s.profileRepo.FindProfileByUserID(userID)
	if err != nil {
		return nil, errors.New("profil tidak ditemukan, silakan buat profil terlebih dahulu")
	}

	exp := &models.UserExperience{
		UserProfileID: profile.ID,
		CompanyName:   req.CompanyName,
		Position:      req.Position,
		StartYear:     req.StartYear,
		EndYear:       req.EndYear,
		Description:   req.Description,
	}

	if err := s.profileRepo.AddExperience(exp); err != nil {
		return nil, errors.New("gagal menambahkan pengalaman")
	}
	return exp, nil
}

func (s *profileService) UpdateExperience(userID uint, expID uint, req ExperienceRequest) (*models.UserExperience, error) {
	profile, err := s.profileRepo.FindProfileByUserID(userID)
	if err != nil {
		return nil, errors.New("profil tidak ditemukan")
	}

	exp, err := s.profileRepo.FindExperienceByID(expID)
	if err != nil {
		return nil, errors.New("pengalaman tidak ditemukan")
	}

	if exp.UserProfileID != profile.ID {
		return nil, errors.New("akses ditolak")
	}

	if req.CompanyName != "" {
		exp.CompanyName = req.CompanyName
	}
	if req.Position != "" {
		exp.Position = req.Position
	}
	if req.StartYear != 0 {
		exp.StartYear = req.StartYear
	}
	exp.EndYear = req.EndYear
	exp.Description = req.Description

	if err := s.profileRepo.UpdateExperience(exp); err != nil {
		return nil, errors.New("gagal memperbarui pengalaman")
	}
	return exp, nil
}

func (s *profileService) DeleteExperience(userID uint, expID uint) error {
	profile, err := s.profileRepo.FindProfileByUserID(userID)
	if err != nil {
		return errors.New("profil tidak ditemukan")
	}

	exp, err := s.profileRepo.FindExperienceByID(expID)
	if err != nil {
		return errors.New("pengalaman tidak ditemukan")
	}

	if exp.UserProfileID != profile.ID {
		return errors.New("akses ditolak")
	}

	return s.profileRepo.DeleteExperience(expID)
}

func (s *profileService) UpdateProfilePicture(userID uint, pictureURL string) error {
	profile, err := s.profileRepo.FindProfileByUserID(userID)
	if err != nil {
		return errors.New("profil tidak ditemukan")
	}
	profile.ProfilePicture = pictureURL
	return s.profileRepo.UpdateProfile(profile)
}
