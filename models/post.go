package models

import "gorm.io/gorm"

type Post struct {
	gorm.Model
	UserID    uint        `gorm:"not null"`
	Category  *string     `gorm:"default:null"`
	Title     string      `gorm:"not null"`
	Content   string      `gorm:"not null"`
	ImageURL  *string     `gorm:"default:null"` // Cloudinary URL (deprecated, kept for backward compatibility)
	User      User        `gorm:"foreignKey:UserID"`
	Images    []PostImage `gorm:"foreignKey:PostID"`
	Comments  []Comment   `gorm:"foreignKey:PostID"`
	Reactions []Reaction  `gorm:"foreignKey:PostID"`
	Votes     []Vote      `gorm:"foreignKey:PostID"`
}
