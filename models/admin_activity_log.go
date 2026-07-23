package models

import "time"

type AdminActivityLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	AdminID    uint      `gorm:"not null;index" json:"admin_id"`
	Admin      User      `gorm:"foreignKey:AdminID" json:"admin"`
	Action     string    `gorm:"type:varchar(50);not null;index" json:"action"`
	TargetType string    `gorm:"type:varchar(50);not null;index" json:"target_type"`
	TargetID   uint      `gorm:"not null;index" json:"target_id"`
	Details    *string   `gorm:"type:text" json:"details"`
	IPAddress  *string   `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent  *string   `gorm:"type:varchar(255)" json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`
}
