package models

import "gorm.io/gorm"

type GroupArticle struct {
	gorm.Model
	GroupID   uint            `gorm:"not null"`
	UserID    uint            `gorm:"not null"`
	Title     string          `gorm:"not null"`
	Content   string          `gorm:"not null"`
	MediaURL  *string         `gorm:"default:null"`
	User      User            `gorm:"foreignKey:UserID"`
	Comments  []GroupComment  `gorm:"foreignKey:ArticleID"`
	Reactions []GroupReaction `gorm:"foreignKey:ArticleID"`
}
