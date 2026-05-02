package models

import "gorm.io/gorm"

// Notification stores a user-facing notification event.
// Delivered in real-time via WebSocket if the user is connected,
// and persisted for offline retrieval via REST.
type Notification struct {
	gorm.Model
	UserID           uint   `gorm:"not null;index:idx_notif_unread,priority:1;index:idx_notif_throttle,priority:1"` // recipient
	NotificationType string `gorm:"not null;index:idx_notif_throttle,priority:2"`       // e.g. "new_follower"
	Title            string `gorm:"not null"`
	Body             string `gorm:"not null"`
	ReferenceType    string `gorm:"not null;default:'';index:idx_notif_throttle,priority:3"` // "follow"|"mentor_request"|"post"|"job_application"|"message"|"report"|"group"|"event"
	ReferenceID      uint   `gorm:"not null;default:0;index:idx_notif_throttle,priority:4"`
	RedirectURL      string `gorm:"not null;default:''"` // frontend route hint for click navigation
	IsRead           bool   `gorm:"not null;default:false;index:idx_notif_unread,priority:2"`

	User             User   `gorm:"foreignKey:UserID"`
}
