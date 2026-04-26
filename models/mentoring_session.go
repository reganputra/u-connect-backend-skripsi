package models

import (
	"time"

	"gorm.io/gorm"
)

// MentoringSession is a scheduled session between an approved mentor-mentee pair.
type MentoringSession struct {
	gorm.Model
	RequestID   uint          `gorm:"not null;index"` // FK → mentor_requests (must be approved)
	Request     MentorRequest `gorm:"foreignKey:RequestID"`
	MentorID    uint          `gorm:"not null;index"`
	Mentor      User          `gorm:"foreignKey:MentorID"`
	StudentID   uint          `gorm:"not null;index;index:idx_mentor_sess_student,priority:1"`
	Student     User          `gorm:"foreignKey:StudentID"`
	Topic       string        `gorm:"not null"`
	Notes       *string       `gorm:"default:null"`
	SessionDate *time.Time    `gorm:"default:null"`
	Status      string        `gorm:"not null;default:'scheduled';index:idx_mentor_sess_student,priority:2"` // scheduled | completed | cancelled
	CompletedAt *time.Time    `gorm:"default:null"`
	CancelledAt *time.Time    `gorm:"default:null"`
}
