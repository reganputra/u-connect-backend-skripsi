package models

import "gorm.io/gorm"

// Vote is polymorphic: either PostID or CommentID is set (not both)
type Vote struct {
	gorm.Model
	UserID    uint  `gorm:"not null"`
	PostID    *uint `gorm:"default:null"`
	CommentID *uint `gorm:"default:null"`
	Value     int   `gorm:"not null"` // 1 = upvote, -1 = downvote
}
