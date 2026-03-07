package models

import "gorm.io/gorm"

type GroupComment struct {
	gorm.Model
	ArticleID       uint            `gorm:"not null"`
	UserID          uint            `gorm:"not null"`
	Content         string          `gorm:"not null"`
	ParentCommentID *uint           `gorm:"default:null"`
	User            User            `gorm:"foreignKey:UserID"`
	Replies         []GroupComment  `gorm:"foreignKey:ParentCommentID"`
	Reactions       []GroupReaction `gorm:"foreignKey:CommentID"`
}
