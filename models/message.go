package models

import "gorm.io/gorm"

// Message represents a private message between two users.
// Messages are immutable — no edit or delete after creation.
type Message struct {
	gorm.Model
	SenderID   uint   `gorm:"not null;index"`
	ReceiverID uint   `gorm:"not null;index"`
	Content    string `gorm:"not null"`
	IsRead     bool   `gorm:"not null;default:false"`
	Sender     User   `gorm:"foreignKey:SenderID"`
	Receiver   User   `gorm:"foreignKey:ReceiverID"`
}
