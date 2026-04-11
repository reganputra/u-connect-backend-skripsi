package models

import "gorm.io/gorm"

type PortfolioItem struct {
	gorm.Model
	UserID      uint    `gorm:"not null"` // FK → users.id
	Title       string  `gorm:"not null"`
	Description *string `gorm:"default:null"`
	Category    *string `gorm:"default:null"`
	Tags        *string `gorm:"default:null"` // comma-separated
	StartDate   *string `gorm:"default:null"` // YYYY-MM
	EndDate     *string `gorm:"default:null"` // YYYY-MM
	MediaURL    *string `gorm:"default:null"` // Cloudinary URL
	Link        *string `gorm:"default:null"` // External URL (optional, non-Cloudinary)
}
