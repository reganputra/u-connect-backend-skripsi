package models

import "gorm.io/gorm"

type Group struct {
	gorm.Model
	OwnerID     uint           `gorm:"not null"`
	Category    string         `gorm:"not null"`
	Title       string         `gorm:"not null"`
	Description *string        `gorm:"default:null"`
	Rules       *string        `gorm:"default:null"`
	BannerURL   *string        `gorm:"default:null"`
	Owner       User           `gorm:"foreignKey:OwnerID"`
	Members     []GroupMember  `gorm:"foreignKey:GroupID"`
	Articles    []GroupArticle `gorm:"foreignKey:GroupID"`
}
