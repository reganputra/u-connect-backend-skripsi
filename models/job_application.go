package models

import "gorm.io/gorm"

type JobApplication struct {
	gorm.Model
	JobID       uint `gorm:"not null;uniqueIndex:idx_job_user"`
	Job         Job  `gorm:"foreignKey:JobID"`
	UserID      uint `gorm:"not null;uniqueIndex:idx_job_user"`
	User        User `gorm:"foreignKey:UserID"`
	CoverLetter *string
	ResumeURL   string `gorm:"not null"`
	Status      string `gorm:"default:'pending'"` // pending, reviewed, accepted, rejected, withdrawn
}
