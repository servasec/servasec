package models

import (
	"time"
)

type UserApiKey struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"not null;index" json:"userId"`
	Name       string     `gorm:"not null;size:100" json:"name"`
	KeyHash    string     `gorm:"not null;uniqueIndex;size:64" json:"-"`
	KeyPrefix  string     `gorm:"not null;size:20" json:"keyPrefix"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	ExpiresAt  *time.Time `json:"expiresAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}
