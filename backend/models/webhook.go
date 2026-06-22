package models

import "time"

type Webhook struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ApplicationID uint      `gorm:"not null;index" json:"applicationId"`
	URL           string    `gorm:"size:500;not null" json:"url"`
	Secret        string    `gorm:"size:64" json:"secret"`
	Events        string    `gorm:"type:text" json:"events"`
	IsActive      bool      `gorm:"default:true" json:"isActive"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
