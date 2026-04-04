package models

import (
	"gorm.io/gorm"
	"time"
)

// MentoringSession is a scheduled session between an approved mentor-mentee pair.
type MentoringSession struct {
	gorm.Model
	RequestID   uint          `gorm:"not null;index"` // FK → mentor_requests (must be approved)
	Request     MentorRequest `gorm:"foreignKey:RequestID"`
	MentorID    uint          `gorm:"not null;index"`
	Mentor      User          `gorm:"foreignKey:MentorID"`
	StudentID   uint          `gorm:"not null;index"`
	Student     User          `gorm:"foreignKey:StudentID"`
	Topic       string        `gorm:"not null"`
	Notes       *string       `gorm:"default:null"`
	SessionDate *time.Time    `gorm:"default:null"`
	Status      string        `gorm:"not null;default:'scheduled'"` // scheduled | completed | cancelled
}
