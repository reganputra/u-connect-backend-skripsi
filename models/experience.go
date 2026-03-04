package models

import "gorm.io/gorm"

type UserExperience struct {
	gorm.Model
	UserProfileID uint    `gorm:"not null"` // FK → user_profiles.id
	CompanyName   string  `gorm:"not null"`
	Position      string  `gorm:"not null"`
	StartYear     int     `gorm:"not null"`
	EndYear       *int    `gorm:"default:null"` // null = currently working here
	Description   *string `gorm:"default:null"`
}
