package models

import "gorm.io/gorm"

// Reaction is polymorphic: either PostID or CommentID is set (not both)
type Reaction struct {
	gorm.Model
	UserID    uint   `gorm:"not null"`
	PostID    *uint  `gorm:"default:null"`
	CommentID *uint  `gorm:"default:null"`
	Type      string `gorm:"not null"` // like|love|haha|wow|sad|angry
}
