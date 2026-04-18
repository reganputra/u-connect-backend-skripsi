package models

import "gorm.io/gorm"

type GroupArticle struct {
	gorm.Model
	GroupID   uint                `gorm:"not null"`
	UserID    uint                `gorm:"not null"`
	Title     string              `gorm:"not null"`
	Content   string              `gorm:"not null"`
	MediaURL  *string             `gorm:"default:null"` // Deprecated: keeping for backward compatibility
	User      User                `gorm:"foreignKey:UserID"`
	Medias    []GroupArticleImage `gorm:"foreignKey:ArticleID"`
	Comments  []GroupComment      `gorm:"foreignKey:ArticleID"`
	Reactions []GroupReaction     `gorm:"foreignKey:ArticleID"`
}
