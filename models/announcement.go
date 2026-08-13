package models

import "time"

type Announcement struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	AdminID    uint       `gorm:"not null;index" json:"admin_id"`
	Admin      User       `gorm:"foreignKey:AdminID" json:"admin"`
	Title      string     `gorm:"type:varchar(150);not null" json:"title"`
	Message    string     `gorm:"type:text;not null" json:"message"`
	ActionURL  *string    `gorm:"type:varchar(255)" json:"action_url"`
	ActionText *string    `gorm:"type:varchar(50)" json:"action_text"`
	TargetRole string     `gorm:"type:varchar(20);not null;default:'all';index" json:"target_role"` // all | student | alumni | partner
	Priority   string     `gorm:"type:varchar(20);not null;default:'info'" json:"priority"`        // info | warning | urgent
	IsBanner   bool       `gorm:"not null;default:false;index" json:"is_banner"`
	ExpiresAt  *time.Time `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
}
