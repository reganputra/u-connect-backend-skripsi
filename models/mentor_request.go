package models

import (
	"time"

	"gorm.io/gorm"
)

// AfterMigrate is called automatically by GORM after AutoMigrate processes
// the mentor_requests table. It creates a partial unique index that enforces
// the rule: a student can only have one active (pending OR approved) request
// per mentor at a time.
//
// We cannot use GORM's struct-tag where: clause here because it does not
// support IN (...) predicates and produces truncated SQL.
func (MentorRequest) AfterMigrate(tx *gorm.DB) error {
	return tx.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_active_request_pair
		ON mentor_requests (mentor_id, student_id)
		WHERE deleted_at IS NULL AND status IN ('pending', 'approved')
	`).Error
}

// MentorRequest represents a student's request to be mentored by an alumni.
// Status flow: pending → approved | rejected | withdrawn
type MentorRequest struct {
	gorm.Model
	MentorID        uint       `gorm:"not null;index"`
	Mentor          User       `gorm:"foreignKey:MentorID"`
	StudentID       uint       `gorm:"not null;index"`
	Student         User       `gorm:"foreignKey:StudentID"`
	Message         *string    `gorm:"default:null"`                                                                                                                  // optional message from student
	Status          string     `gorm:"not null;default:'pending'"` // pending | approved | rejected | withdrawn
	RejectReason    *string    `gorm:"default:null"`
	SimilarityScore *float64   `gorm:"default:null"` // cosine similarity score recorded at request time
	ApprovedAt      *time.Time `gorm:"default:null"`
	RejectedAt      *time.Time `gorm:"default:null"`
	WithdrawnAt     *time.Time `gorm:"default:null"`
}
