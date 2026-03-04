package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name     string `gorm:"not null"`
	Email    string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
	Role     string `gorm:"not null"` // alumni | student | partner

	// alumni & student only
	Faculty    *string `gorm:"default:null"`
	Major      *string `gorm:"default:null"`
	YearEnroll *int    `gorm:"default:null"`

	// partner only
	CompanyName *string `gorm:"default:null"`
}
