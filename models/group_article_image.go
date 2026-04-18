package models

import "gorm.io/gorm"

type GroupArticleImage struct {
	gorm.Model
	ArticleID uint         `gorm:"not null;index"`
	ImageURL  string       `gorm:"not null"` // Cloudinary URL
	Article   GroupArticle `gorm:"foreignKey:ArticleID"`
}
