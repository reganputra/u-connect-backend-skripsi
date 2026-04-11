package models

import "gorm.io/gorm"

type CompanyProfile struct {
	gorm.Model
	CompanyName  string  `gorm:"not null"`
	Description  *string `gorm:"default:null"`
	IndustryType *string `gorm:"default:null"`
	Location     *string `gorm:"default:null"`
	EmployeeSize *int    `gorm:"default:null"` // must be >= 0
	WebsiteURL   *string `gorm:"default:null"`
}
