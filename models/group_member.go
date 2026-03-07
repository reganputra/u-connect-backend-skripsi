package models

import "gorm.io/gorm"

type GroupMember struct {
	gorm.Model
	GroupID uint   `gorm:"not null;uniqueIndex:idx_group_user"`
	UserID  uint   `gorm:"not null;uniqueIndex:idx_group_user"`
	Role    string `gorm:"not null;default:'member'"` // owner|member
	User    User   `gorm:"foreignKey:UserID"`
}
