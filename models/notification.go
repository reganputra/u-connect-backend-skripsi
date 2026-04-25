package models

import "gorm.io/gorm"

// Notification stores a user-facing notification event.
// Delivered in real-time via WebSocket if the user is connected,
// and persisted for offline retrieval via REST.
type Notification struct {
	gorm.Model
	UserID           uint   `gorm:"not null;index"` // recipient
	NotificationType string `gorm:"not null"`       // e.g. "new_follower"
	Title            string `gorm:"not null"`
	Body             string `gorm:"not null"`
	ReferenceType    string `gorm:"not null;default:''"` // "follow"|"mentor_request"|"post"|"job_application"|"message"|"report"|"group"|"event"
	ReferenceID      uint   `gorm:"not null;default:0"`
	RedirectURL      string `gorm:"not null;default:''"` // frontend route hint for click navigation
	IsRead           bool   `gorm:"not null;default:false"`
	User             User   `gorm:"foreignKey:UserID"`
}
