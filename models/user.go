package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name       string  `gorm:"not null"`
	Email      string  `gorm:"uniqueIndex;not null"`
	Password   string  `gorm:"not null" json:"-"`
	Role       string  `gorm:"not null"` // alumni | student | partner | admin
	IsActive   bool    `gorm:"default:true"`
	PictureURL *string `gorm:"-" json:"picture_url"`

	// alumni & student only
	Faculty    *string `gorm:"default:null"`
	Major      *string `gorm:"default:null"`
	YearEnroll *int    `gorm:"default:null"`

	// partner only
	CompanyName *string `gorm:"default:null"`

	// Forgot password rate limiting
	ResetAttempts  int        `gorm:"default:0"`
	ResetLockedAt  *time.Time `gorm:"default:null"`

	// Association — populated only when explicitly Preloaded
	Profile *UserProfile `gorm:"foreignKey:UserID"`
}
