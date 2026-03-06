package models

import "gorm.io/gorm"

type Comment struct {
	gorm.Model
	PostID          uint       `gorm:"not null"`
	UserID          uint       `gorm:"not null"`
	Content         string     `gorm:"not null"`
	ParentCommentID *uint      `gorm:"default:null"` // null = top-level, set = reply
	User            User       `gorm:"foreignKey:UserID"`
	Replies         []Comment  `gorm:"foreignKey:ParentCommentID"`
	Reactions       []Reaction `gorm:"foreignKey:CommentID"`
	Votes           []Vote     `gorm:"foreignKey:CommentID"`
}
