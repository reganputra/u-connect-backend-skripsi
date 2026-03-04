package service

import (
	"errors"

	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
)

var validJobStatuses = map[string]bool{
	"employed":         true,
	"entrepreneur":     true,
	"continuing_study": true,
	"unemployed":       true,
	"freelance":        true,
	"student":          true,
}

type ProfileRequest struct {
	Bio      *string `json:"bio"`
	Location *string `json:"location"`

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

func validateJobStatus(req ProfileRequest) error {
	if req.JobStatus == nil {
		return nil
	}
	status := *req.JobStatus
	if !validJobStatuses[status] {
		return errors.New("invalid job_status: must be employed, entrepreneur, continuing_study, unemployed, freelance, or student")
	}
	switch status {
	case "employed":
		if req.Position == nil || *req.Position == "" {
			return errors.New("position is required when job_status is employed")
		}
		if req.CompanyName == nil || *req.CompanyName == "" {
			return errors.New("company_name is required when job_status is employed")
		}
	case "entrepreneur":
		if req.IndustryName == nil || *req.IndustryName == "" {
			return errors.New("industry_name is required when job_status is entrepreneur")
		}
	case "continuing_study":
		if req.EducationalLevel == nil || *req.EducationalLevel == "" {
			return errors.New("educational_level is required when job_status is continuing_study")
		}
		if req.AdvancedStudyProgram == nil || *req.AdvancedStudyProgram == "" {
			return errors.New("advanced_study_program is required when job_status is continuing_study")
		}
		if req.InstitutionName == nil || *req.InstitutionName == "" {
			return errors.New("institution_name is required when job_status is continuing_study")
		}
	case "unemployed", "freelance":
		if req.StatusDescription == nil || *req.StatusDescription == "" {
			return errors.New("status_description is required when job_status is unemployed or freelance")
		}
	}
	return nil
}

func nullFieldsByStatus(profile *models.UserProfile) {
	if profile.JobStatus == nil {
		return
	}
	status := *profile.JobStatus
	// Clear fields irrelevant to the active status
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

func CreateProfile(userID uint, req ProfileRequest) (*models.UserProfile, error) {
	// Check for duplicate
	existing, _ := repository.FindProfileByUserID(userID)
	if existing != nil {
		return nil, errors.New("profile already exists")
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

	nullFieldsByStatus(profile)

	if err := repository.CreateProfile(profile); err != nil {
		return nil, errors.New("failed to create profile")
	}
	return profile, nil
}

func GetProfile(userID uint) (*models.UserProfile, error) {
	profile, err := repository.FindProfileByUserID(userID)
	if err != nil {
		return nil, errors.New("profile not found")
	}
	return profile, nil
}

func UpdateProfile(userID uint, req ProfileRequest) (*models.UserProfile, error) {
	profile, err := repository.FindProfileByUserID(userID)
	if err != nil {
		return nil, errors.New("profile not found")
	}

	// Apply updates for non-nil fields only
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
	if req.StatusDescription != nil {
		profile.StatusDescription = req.StatusDescription
	}

	if err := validateJobStatus(req); err != nil {
		return nil, err
	}
	nullFieldsByStatus(profile)

	if err := repository.UpdateProfile(profile); err != nil {
		return nil, errors.New("failed to update profile")
	}
	return profile, nil
}

func DeleteProfile(userID uint) error {
	_, err := repository.FindProfileByUserID(userID)
	if err != nil {
		return errors.New("profile not found")
	}
	return repository.DeleteProfileByUserID(userID)
}

func AddExperience(userID uint, req ExperienceRequest) (*models.UserExperience, error) {
	if req.CompanyName == "" || req.Position == "" || req.StartYear == 0 {
		return nil, errors.New("company_name, position, and start_year are required")
	}

	profile, err := repository.FindProfileByUserID(userID)
	if err != nil {
		return nil, errors.New("profile not found, please create a profile first")
	}

	exp := &models.UserExperience{
		UserProfileID: profile.ID,
		CompanyName:   req.CompanyName,
		Position:      req.Position,
		StartYear:     req.StartYear,
		EndYear:       req.EndYear,
		Description:   req.Description,
	}

	if err := repository.AddExperience(exp); err != nil {
		return nil, errors.New("failed to add experience")
	}
	return exp, nil
}

func UpdateExperience(userID uint, expID uint, req ExperienceRequest) (*models.UserExperience, error) {
	profile, err := repository.FindProfileByUserID(userID)
	if err != nil {
		return nil, errors.New("profile not found")
	}

	exp, err := repository.FindExperienceByID(expID)
	if err != nil {
		return nil, errors.New("experience not found")
	}

	// Verify ownership
	if exp.UserProfileID != profile.ID {
		return nil, errors.New("access denied")
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

	if err := repository.UpdateExperience(exp); err != nil {
		return nil, errors.New("failed to update experience")
	}
	return exp, nil
}

func DeleteExperience(userID uint, expID uint) error {
	profile, err := repository.FindProfileByUserID(userID)
	if err != nil {
		return errors.New("profile not found")
	}

	exp, err := repository.FindExperienceByID(expID)
	if err != nil {
		return errors.New("experience not found")
	}

	if exp.UserProfileID != profile.ID {
		return errors.New("access denied")
	}

	return repository.DeleteExperience(expID)
}

func UpdateProfilePicture(userID uint, pictureURL string) error {
	profile, err := repository.FindProfileByUserID(userID)
	if err != nil {
		return errors.New("profile not found")
	}
	profile.ProfilePicture = pictureURL
	return repository.UpdateProfile(profile)
}
