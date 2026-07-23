package repository

import (
	"time"

	"github.com/reganputra/skripsi-backend/constant"
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type MentoringSessionRepository interface {
	Create(session *models.MentoringSession) error
	FindByID(id uint) (*models.MentoringSession, error)
	FindByMentorID(mentorID uint) ([]models.MentoringSession, error)
	FindByStudentID(studentID uint) ([]models.MentoringSession, error)
	Update(session *models.MentoringSession) error
	CancelScheduledByPair(mentorID, studentID uint) (int64, error)
	// FindApprovedRequest checks that an approved mentor-student relationship exists.
	FindApprovedRequest(mentorID, studentID uint) (*models.MentorRequest, error)
}

type mentoringSessionRepository struct {
	db *gorm.DB
}

func NewMentoringSessionRepository(db *gorm.DB) MentoringSessionRepository {
	return &mentoringSessionRepository{db: db}
}

func (r *mentoringSessionRepository) Create(session *models.MentoringSession) error {
	return r.db.Create(session).Error
}

func (r *mentoringSessionRepository) FindByID(id uint) (*models.MentoringSession, error) {
	var s models.MentoringSession
	err := r.db.
		Preload("Mentor").
		Preload("Student").
		First(&s, id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *mentoringSessionRepository) FindByMentorID(mentorID uint) ([]models.MentoringSession, error) {
	var sessions []models.MentoringSession
	err := r.db.
		Preload("Student").
		Where("mentor_id = ?", mentorID).
		Order("created_at DESC").
		Find(&sessions).Error
	return sessions, err
}

func (r *mentoringSessionRepository) FindByStudentID(studentID uint) ([]models.MentoringSession, error) {
	var sessions []models.MentoringSession
	err := r.db.
		Preload("Mentor").
		Where("student_id = ?", studentID).
		Order("created_at DESC").
		Find(&sessions).Error
	return sessions, err
}

func (r *mentoringSessionRepository) Update(session *models.MentoringSession) error {
	return r.db.Save(session).Error
}

func (r *mentoringSessionRepository) CancelScheduledByPair(mentorID, studentID uint) (int64, error) {
	now := time.Now()
	result := r.db.Model(&models.MentoringSession{}).
		Where("mentor_id = ? AND student_id = ? AND status = 'scheduled'", mentorID, studentID).
		Updates(map[string]any{
			"status":       constant.StatusCancelled,
			"cancelled_at": now,
		})
	return result.RowsAffected, result.Error
}

func (r *mentoringSessionRepository) FindApprovedRequest(mentorID, studentID uint) (*models.MentorRequest, error) {
	var req models.MentorRequest
	err := r.db.
		Where("mentor_id = ? AND student_id = ? AND status = 'approved'", mentorID, studentID).
		First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}
