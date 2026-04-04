package models

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	Name        string  `gorm:"uniqueIndex;not null"`
	Description *string `gorm:"default:null"`
}
