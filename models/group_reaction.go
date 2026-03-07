package models

import "gorm.io/gorm"

// GroupReaction is polymorphic: either ArticleID or CommentID is set (not both)
type GroupReaction struct {
	gorm.Model
	UserID    uint   `gorm:"not null"`
	ArticleID *uint  `gorm:"default:null"`
	CommentID *uint  `gorm:"default:null"`
	Type      string `gorm:"not null"` // like|love|haha|wow|sad|angry
}
