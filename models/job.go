package models

import "gorm.io/gorm"

type Job struct {
	gorm.Model
	UserID      uint   `gorm:"not null"`
	User        User   `gorm:"foreignKey:UserID"`
	Title       string `gorm:"not null"`
	Description *string
	CompanyName string `gorm:"not null"`
	Location    *string
	JobType     string `gorm:"not null"`       // full-time, part-time, internship, contract, freelance
	Status      string `gorm:"default:'open'"` // open, closed, filled
	SalaryRange *string
	ImageURL    *string

	Applications []JobApplication `gorm:"foreignKey:JobID"`
}
