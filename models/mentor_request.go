package models

import "gorm.io/gorm"

// MentorRequest represents a student's request to be mentored by an alumni.
// Status flow: pending → approved | rejected
type MentorRequest struct {
	gorm.Model
	MentorID        uint     `gorm:"not null;index"`
	Mentor          User     `gorm:"foreignKey:MentorID"`
	StudentID       uint     `gorm:"not null;index"`
	Student         User     `gorm:"foreignKey:StudentID"`
	Message         *string  `gorm:"default:null"`                // optional message from student
	Status          string   `gorm:"not null;default:'pending'"` // pending | approved | rejected
	RejectReason    *string  `gorm:"default:null"`
	SimilarityScore *float64 `gorm:"default:null"` // cosine similarity score recorded at request time
}
