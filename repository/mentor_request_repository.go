package repository

import (
	"errors"
	"time"

	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrRequestNotFound       = errors.New("request_not_found")
	ErrRequestForbidden      = errors.New("request_forbidden")
	ErrRequestAlreadyHandled = errors.New("request_already_handled")
	ErrStudentLimitReached   = errors.New("student_limit_reached")
	ErrMentorProfileMissing  = errors.New("mentor_profile_missing")
	ErrMentorCapacityFull    = errors.New("mentor_capacity_full")
)

type MentorRequestRepository interface {
	Create(req *models.MentorRequest) error
	FindByID(id uint) (*models.MentorRequest, error)
	FindByMentorID(mentorID uint) ([]models.MentorRequest, error)
	FindByStudentID(studentID uint) ([]models.MentorRequest, error)
	// FindActiveDuplicate returns a pending or approved request between same pair (for duplicate check).
	FindActiveDuplicate(studentID, mentorID uint) (*models.MentorRequest, error)
	// CountApprovedForStudent counts how many mentors a student currently has (approved).
	CountApprovedForStudent(studentID uint) (int64, error)
	ApprovePendingTransactional(mentorID, requestID uint) (*models.MentorRequest, error)
	EndMentorshipTransactional(mentorID, requestID uint, reason *string) (*models.MentorRequest, int64, error)
	Update(req *models.MentorRequest) error
	// FindGroundTruth returns all (student_id, mentor_id) request pairs for MAP evaluation.
	FindGroundTruth() ([]GroundTruthEntry, error)
}

// GroundTruthEntry is a flat row from mentor_requests used to build the MAP evaluation ground truth.
type GroundTruthEntry struct {
	StudentID uint `gorm:"column:student_id"`
	MentorID  uint `gorm:"column:mentor_id"`
}

type mentorRequestRepository struct {
	db *gorm.DB
}

func NewMentorRequestRepository(db *gorm.DB) MentorRequestRepository {
	return &mentorRequestRepository{db: db}
}

func (r *mentorRequestRepository) Create(req *models.MentorRequest) error {
	return r.db.Create(req).Error
}

func (r *mentorRequestRepository) FindByID(id uint) (*models.MentorRequest, error) {
	var req models.MentorRequest
	err := r.db.
		Preload("Mentor").
		Preload("Student").
		First(&req, id).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *mentorRequestRepository) FindByMentorID(mentorID uint) ([]models.MentorRequest, error) {
	var reqs []models.MentorRequest
	err := r.db.
		Preload("Mentor").
		Preload("Mentor.Profile").
		Preload("Student").
		Preload("Student.Profile").
		Where("mentor_id = ?", mentorID).
		Order("created_at DESC").
		Find(&reqs).Error
	return reqs, err
}

func (r *mentorRequestRepository) FindByStudentID(studentID uint) ([]models.MentorRequest, error) {
	var reqs []models.MentorRequest
	err := r.db.
		Preload("Mentor").
		Preload("Mentor.Profile").
		Preload("Student").
		Preload("Student.Profile").
		Where("student_id = ?", studentID).
		Order("created_at DESC").
		Find(&reqs).Error
	return reqs, err
}

func (r *mentorRequestRepository) FindActiveDuplicate(studentID, mentorID uint) (*models.MentorRequest, error) {
	var req models.MentorRequest
	err := r.db.
		Where("student_id = ? AND mentor_id = ? AND status IN ('pending','approved')", studentID, mentorID).
		First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *mentorRequestRepository) CountApprovedForStudent(studentID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.MentorRequest{}).
		Where("student_id = ? AND status = 'approved'", studentID).
		Count(&count).Error
	return count, err
}

func (r *mentorRequestRepository) ApprovePendingTransactional(mentorID, requestID uint) (*models.MentorRequest, error) {
	var result models.MentorRequest
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var req models.MentorRequest
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&req, requestID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRequestNotFound
			}
			return err
		}

		if req.MentorID != mentorID {
			return ErrRequestForbidden
		}
		if req.Status != "pending" {
			return ErrRequestAlreadyHandled
		}

		var studentApprovedCount int64
		if err := tx.Model(&models.MentorRequest{}).
			Where("student_id = ? AND status = 'approved'", req.StudentID).
			Count(&studentApprovedCount).Error; err != nil {
			return err
		}
		if studentApprovedCount >= 2 {
			return ErrStudentLimitReached
		}

		var profile struct {
			MentorQuota *int
		}
		if err := tx.Table("user_profiles").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("mentor_quota").
			Where("user_id = ? AND deleted_at IS NULL", mentorID).
			Take(&profile).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrMentorProfileMissing
			}
			return err
		}
		if profile.MentorQuota == nil {
			return ErrMentorProfileMissing
		}

		var mentorApprovedCount int64
		if err := tx.Model(&models.MentorRequest{}).
			Where("mentor_id = ? AND status = 'approved'", mentorID).
			Count(&mentorApprovedCount).Error; err != nil {
			return err
		}
		if mentorApprovedCount >= int64(*profile.MentorQuota) {
			return ErrMentorCapacityFull
		}

		now := time.Now()
		req.Status = "approved"
		req.ApprovedAt = &now
		req.RejectedAt = nil
		req.WithdrawnAt = nil
		if err := tx.Save(&req).Error; err != nil {
			return err
		}

		if err := tx.
			Preload("Mentor").
			Preload("Student").
			First(&result, req.ID).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *mentorRequestRepository) EndMentorshipTransactional(mentorID, requestID uint, reason *string) (*models.MentorRequest, int64, error) {
	var result models.MentorRequest
	var cancelledSessions int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var req models.MentorRequest
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Mentor").
			Preload("Student").
			First(&req, requestID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRequestNotFound
			}
			return err
		}

		if req.MentorID != mentorID {
			return ErrRequestForbidden
		}
		if req.Status != "approved" {
			return ErrRequestAlreadyHandled
		}

		now := time.Now()
		req.Status = "ended"
		req.EndedAt = &now
		req.EndReason = reason
		req.ApprovedAt = req.ApprovedAt
		req.RejectedAt = nil
		req.WithdrawnAt = nil
		if err := tx.Save(&req).Error; err != nil {
			return err
		}

		cancelResult := tx.Model(&models.MentoringSession{}).
			Where("mentor_id = ? AND student_id = ? AND status = 'scheduled'", mentorID, req.StudentID).
			Updates(map[string]any{
				"status":       "cancelled",
				"cancelled_at": now,
			})
		if cancelResult.Error != nil {
			return cancelResult.Error
		}
		cancelledSessions = cancelResult.RowsAffected

		if err := tx.
			Preload("Mentor").
			Preload("Student").
			First(&result, req.ID).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return &result, cancelledSessions, nil
}

func (r *mentorRequestRepository) Update(req *models.MentorRequest) error {
	return r.db.Save(req).Error
}

// FindGroundTruth fetches all (student_id, mentor_id) pairs from mentor_requests.
// Every request a student ever made — regardless of status — is treated as a
// positive relevance signal (the student chose to request that mentor).
func (r *mentorRequestRepository) FindGroundTruth() ([]GroundTruthEntry, error) {
	var rows []GroundTruthEntry
	err := r.db.Table("mentor_requests").
		Select("student_id, mentor_id").
		Where("deleted_at IS NULL").
		Scan(&rows).Error
	return rows, err
}

