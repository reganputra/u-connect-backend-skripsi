package models

import (
	"time"

	"gorm.io/gorm"
)

// Report represents a user-submitted content report.
//
// TargetType : post | group | event | job | comment | group_article
// ReportType : harassment | violence | hate_speech | spam | inappropriate | misinformation | copyright | other
// Status     : pending | resolved | rejected
type Report struct {
	gorm.Model
	ReporterID   uint       `gorm:"not null"`
	Reporter     User       `gorm:"foreignKey:ReporterID"`
	TargetType   string     `gorm:"not null"`
	TargetID     uint       `gorm:"not null"`
	ReportType   string     `gorm:"not null"`
	Description  *string    `gorm:"default:null"` // optional; required when ReportType = "other"
	Status       string     `gorm:"default:'pending'"`
	AdminNote    *string    `gorm:"default:null"`
	ResolvedByID *uint      `gorm:"default:null"`
	ResolvedBy   *User      `gorm:"foreignKey:ResolvedByID"`
	ResolvedAt   *time.Time `gorm:"default:null"`
}
