package models

import "gorm.io/gorm"

type EventRegistration struct {
	gorm.Model
	EventID      uint `gorm:"not null;uniqueIndex:idx_event_user_reg"`
	UserID       uint `gorm:"not null;uniqueIndex:idx_event_user_reg"`
	ReminderSent bool `gorm:"not null;default:false"`
	User         User `gorm:"foreignKey:UserID"`
}
