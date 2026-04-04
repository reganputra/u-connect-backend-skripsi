package models

import "gorm.io/gorm"

// Follow represents a directional follow relationship.
// If A follows B, the record is (FollowerID=A, FollowingID=B).
// Messaging is symmetric: if any follow record exists between two users (in either direction),
// both users may message each other.
type Follow struct {
	gorm.Model
	FollowerID  uint `gorm:"not null;index"`
	FollowingID uint `gorm:"not null;index"`
	Follower    User `gorm:"foreignKey:FollowerID"`
	Following   User `gorm:"foreignKey:FollowingID"`
}
