package models

import "gorm.io/gorm"

// Message represents a private message between two users.
// Messages are immutable — no edit or delete after creation.
type Message struct {
	gorm.Model
	SenderID   uint   `gorm:"not null;index:idx_msg_conv,priority:1"`
	ReceiverID uint   `gorm:"not null;index:idx_msg_conv,priority:2;index:idx_msg_unread,priority:1"`
	Content    string `gorm:"not null"`
	IsRead     bool   `gorm:"not null;default:false;index:idx_msg_unread,priority:2"`
	Sender     User   `gorm:"foreignKey:SenderID"`
	Receiver   User   `gorm:"foreignKey:ReceiverID"`
}
