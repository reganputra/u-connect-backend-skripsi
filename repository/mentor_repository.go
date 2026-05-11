package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

// MentorDoc is a lightweight struct used by the recommendation service for TF-IDF computation.
type MentorDoc struct {
	UserID          uint
	Name            string
	ProfilePicture  string
	MentorBio       string
	Skills          string
	Interests       string
	CareerInterests string
	Position        string
	CompanyName     string
	IndustryName    string
	IndustryType    string
	ExperienceText  string
	YearsExperience int
	MentorQuota     int
}

type MentorRepository interface {
	// FindMentors lists alumni who are registered as mentors and still have capacity.
	FindMentors(page, limit int, search string) ([]models.UserProfile, int64, error)
	// FindMentorProfileByUserID returns a mentor's full profile (includes User & Experiences).
	FindMentorProfileByUserID(userID uint) (*models.UserProfile, error)
	// FindAllMentorDocs returns lightweight mentor data for TF-IDF recommendation.
	FindAllMentorDocs() ([]MentorDoc, error)
	// CountActiveMentees counts how many students are currently approved for a given mentor.
	CountActiveMentees(mentorUserID uint) (int64, error)
}

type mentorRepository struct {
	db *gorm.DB
}

func NewMentorRepository(db *gorm.DB) MentorRepository {
	return &mentorRepository{db: db}
}

// activeMenteesSubquery is used in GORM Where() calls where GORM keeps 'user_profiles' unaliased.
const activeMenteesSubquery = `(
	SELECT COUNT(*) FROM mentor_requests mr
	WHERE mr.mentor_id = user_profiles.user_id
	  AND mr.status = 'approved'
	  AND mr.deleted_at IS NULL
)`

// activeMenteesSubqueryAliased is used inside raw SQL where user_profiles is aliased as 'up'.
const activeMenteesSubqueryAliased = `(
	SELECT COUNT(*) FROM mentor_requests mr
	WHERE mr.mentor_id = up.user_id
	  AND mr.status = 'approved'
	  AND mr.deleted_at IS NULL
)`

func (r *mentorRepository) FindMentors(page, limit int, search string) ([]models.UserProfile, int64, error) {
	var profiles []models.UserProfile
	var total int64

	q := r.db.
		Joins("JOIN users u ON u.id = user_profiles.user_id AND u.deleted_at IS NULL").
		Where("u.role = ?", "alumni").
		Where("user_profiles.mentor_quota IS NOT NULL").
		Where(activeMenteesSubquery + " < user_profiles.mentor_quota")

	if search != "" {
		like := "%" + search + "%"
		q = q.Where("u.name ILIKE ? OR user_profiles.mentor_description ILIKE ? OR user_profiles.skills ILIKE ?",
			like, like, like)
	}

	q.Model(&models.UserProfile{}).Count(&total)

	offset := (page - 1) * limit
	err := q.
		Preload("User").
		Preload("Experiences").
		Offset(offset).Limit(limit).
		Order("user_profiles.created_at DESC").
		Find(&profiles).Error

	return profiles, total, err
}

func (r *mentorRepository) FindMentorProfileByUserID(userID uint) (*models.UserProfile, error) {
	var profile models.UserProfile
	err := r.db.
		Joins("JOIN users u ON u.id = user_profiles.user_id AND u.deleted_at IS NULL").
		Where("user_profiles.user_id = ? AND u.role = ? AND user_profiles.mentor_quota IS NOT NULL", userID, "alumni").
		Preload("User").
		Preload("Experiences").
		First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *mentorRepository) FindAllMentorDocs() ([]MentorDoc, error) {
	var docs []MentorDoc
	err := r.db.Raw(`
		WITH exp AS (
			SELECT
				ue.user_profile_id,
				COALESCE(
					STRING_AGG(
						TRIM(
							COALESCE(ue.company_name, '') || ' ' ||
							COALESCE(ue.position, '') || ' ' ||
							COALESCE(ue.description, '')
						),
						' '
					),
					''
				) AS experience_text,
				COALESCE(
					MAX(COALESCE(ue.end_year, EXTRACT(YEAR FROM NOW())::int)) - MIN(ue.start_year) + 1,
					0
				) AS years_experience
			FROM user_experiences ue
			WHERE ue.deleted_at IS NULL
			GROUP BY ue.user_profile_id
		)
		SELECT
			u.id         AS user_id,
			u.name,
			up.profile_picture,
			COALESCE(up.mentor_description, '')  AS mentor_bio,
			COALESCE(up.skills, '')              AS skills,
			COALESCE(up.interests, '')           AS interests,
			COALESCE(up.career_interests, '')    AS career_interests,
			COALESCE(up.position, '')            AS position,
			COALESCE(up.company_name, '')        AS company_name,
			COALESCE(up.industry_name, '')       AS industry_name,
			COALESCE(up.industry_type, '')       AS industry_type,
			COALESCE(exp.experience_text, '')    AS experience_text,
			COALESCE(exp.years_experience, 0)    AS years_experience,
			up.mentor_quota
		FROM user_profiles up
		JOIN users u ON u.id = up.user_id AND u.deleted_at IS NULL
		LEFT JOIN exp ON exp.user_profile_id = up.id
		WHERE u.role = 'alumni'
		  AND up.mentor_quota IS NOT NULL
		  AND up.deleted_at IS NULL
		  AND ` + activeMenteesSubqueryAliased + ` < up.mentor_quota
	`).Scan(&docs).Error
	return docs, err
}

func (r *mentorRepository) CountActiveMentees(mentorUserID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.MentorRequest{}).
		Where("mentor_id = ? AND status = 'approved'", mentorUserID).
		Count(&count).Error
	return count, err
}
