package models

import "gorm.io/gorm"

type Event struct {
	gorm.Model
	UserID      uint   `gorm:"not null"`
	User        User   `gorm:"foreignKey:UserID"`
	Title       string `gorm:"not null"`
	Description *string
	PhotoURL    *string
	Location    *string
	Capacity    *int
	Status      string `gorm:"default:'upcoming'"`

	Agendas       []EventAgenda       `gorm:"foreignKey:EventID"`
	Registrations []EventRegistration `gorm:"foreignKey:EventID"`
}
