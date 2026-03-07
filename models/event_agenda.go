package models

import (
	"time"

	"gorm.io/gorm"
)

type EventAgenda struct {
	gorm.Model
	EventID     uint       `gorm:"not null"`
	Description string     `gorm:"not null"`
	AgendaTime  *time.Time // optional scheduled time for this agenda item
}
