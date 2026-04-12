package models

import "gorm.io/gorm"

type PostImage struct {
	gorm.Model
	PostID   uint   `gorm:"not null;index"`
	ImageURL string `gorm:"not null"` // Cloudinary URL
	Post     Post   `gorm:"foreignKey:PostID"`
}
