package models

import "time"

type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FindingID uint      `gorm:"not null;index" json:"findingId"`
	UserID    uint      `gorm:"not null;index" json:"userId"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Body      string    `gorm:"type:text;not null" json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}
