package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
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
	Update(req *models.MentorRequest) error
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
		Preload("Student").
		Where("mentor_id = ?", mentorID).
		Order("created_at DESC").
		Find(&reqs).Error
	return reqs, err
}

func (r *mentorRequestRepository) FindByStudentID(studentID uint) ([]models.MentorRequest, error) {
	var reqs []models.MentorRequest
	err := r.db.
		Preload("Mentor").
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

func (r *mentorRequestRepository) Update(req *models.MentorRequest) error {
	return r.db.Save(req).Error
}
