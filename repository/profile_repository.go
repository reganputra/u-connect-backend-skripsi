package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type DirectorySummary struct {
	UserID            uint    `json:"user_id"`
	Name              string  `json:"name"`
	Role              string  `json:"role"`
	ProfilePicture    string  `json:"profile_picture"`
	Bio               *string `json:"bio"`
	Location          *string `json:"location"`
	JobStatus         *string `json:"job_status"`
	Position          *string `json:"position"`
	CompanyName       *string `json:"company_name"`
	Skills            *string `json:"skills"`
	Interests         *string `json:"interests"`
	MentorDescription *string `json:"mentor_description"`
}

type ProfileRepository interface {
	CreateProfile(profile *models.UserProfile) error
	FindProfileByUserID(userID uint) (*models.UserProfile, error)
	EnsureProfileExists(userID uint) error
	BackfillMissingPartnerProfiles() error
	UpdateProfile(profile *models.UserProfile) error
	DeleteProfileByUserID(userID uint) error
	AddExperience(exp *models.UserExperience) error
	FindExperienceByID(id uint) (*models.UserExperience, error)
	UpdateExperience(exp *models.UserExperience) error
	DeleteExperience(id uint) error
	// Directory browsing
	GetAllProfiles(page, limit int) ([]DirectorySummary, int64, error)
	SearchProfiles(query string, page, limit int) ([]DirectorySummary, int64, error)
	GetProfilesByRole(role string, page, limit int) ([]DirectorySummary, int64, error)
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

func (r *profileRepository) EnsureProfileExists(userID uint) error {
	profile := &models.UserProfile{UserID: userID}
	return r.db.Where(models.UserProfile{UserID: userID}).FirstOrCreate(profile).Error
}

func (r *profileRepository) BackfillMissingPartnerProfiles() error {
	var missingUserIDs []uint
	if err := r.db.
		Table("users as u").
		Select("u.id").
		Joins("LEFT JOIN user_profiles up ON up.user_id = u.id").
		Where("u.role = ? AND up.user_id IS NULL", "partner").
		Scan(&missingUserIDs).Error; err != nil {
		return err
	}

	for _, userID := range missingUserIDs {
		if err := r.EnsureProfileExists(userID); err != nil {
			return err
		}
	}

	return nil
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

// GetAllProfiles returns paginated list of all user profiles (newest first).
func (r *profileRepository) GetAllProfiles(page, limit int) ([]DirectorySummary, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var profiles []DirectorySummary
	var total int64

	base := r.db.
		Table("user_profiles as up").
		Select(
			"up.user_id, u.name, u.role, up.profile_picture, up.bio, up.location, "+
				"up.job_status, up.position, up.company_name, up.skills, up.interests, up.mentor_description",
		).
		Joins("JOIN users u ON u.id = up.user_id").
		Where("u.role IN (?, ?, ?) AND u.is_active = true", "student", "alumni", "partner").
		Order("up.created_at DESC")

	base.Count(&total)

	offset := (page - 1) * limit
	err := base.Offset(offset).Limit(limit).Scan(&profiles).Error

	return profiles, total, err
}

// SearchProfiles searches profiles by name, skills, company, or interests.
func (r *profileRepository) SearchProfiles(query string, page, limit int) ([]DirectorySummary, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var profiles []DirectorySummary
	var total int64

	searchPattern := "%" + query + "%"

	base := r.db.
		Table("user_profiles as up").
		Select(
			"up.user_id, u.name, u.role, up.profile_picture, up.bio, up.location, "+
				"up.job_status, up.position, up.company_name, up.skills, up.interests, up.mentor_description",
		).
		Joins("JOIN users u ON u.id = up.user_id").
		Where("u.role IN (?, ?, ?) AND u.is_active = true", "student", "alumni", "partner").
		Where(
			"u.name ILIKE ? OR up.skills ILIKE ? OR up.company_name ILIKE ? OR up.interests ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern,
		).
		Order("up.created_at DESC")

	base.Count(&total)

	offset := (page - 1) * limit
	err := base.Offset(offset).Limit(limit).Scan(&profiles).Error

	return profiles, total, err
}

// GetProfilesByRole returns paginated profiles filtered by user role.
func (r *profileRepository) GetProfilesByRole(role string, page, limit int) ([]DirectorySummary, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Only allow filtering by student, alumni, or partner
	if role != "student" && role != "alumni" && role != "partner" {
		return []DirectorySummary{}, 0, nil
	}

	var profiles []DirectorySummary
	var total int64

	base := r.db.
		Table("user_profiles as up").
		Select(
			"up.user_id, u.name, u.role, up.profile_picture, up.bio, up.location, "+
				"up.job_status, up.position, up.company_name, up.skills, up.interests, up.mentor_description",
		).
		Joins("JOIN users u ON u.id = up.user_id").
		Where("u.role = ? AND u.is_active = true", role).
		Order("up.created_at DESC")

	base.Count(&total)

	offset := (page - 1) * limit
	err := base.Offset(offset).Limit(limit).Scan(&profiles).Error

	return profiles, total, err
}
