package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	Username          string         `gorm:"unique;not null;size:32" json:"username"`
	Email             string         `gorm:"unique;not null;size:254" json:"email"`
	Password          string         `json:"-"`
	Role              string         `gorm:"not null;default:'member'" json:"role"`
	Banned            bool           `json:"banned"`
	OAuthProvider     string         `gorm:"column:oauth_provider;size:50;index:idx_oauth_provider_id" json:"oauthProvider,omitempty"`
	OAuthID           string         `gorm:"column:oauth_id;size:255;index:idx_oauth_provider_id" json:"oauthId,omitempty"`
	AvatarURL         string         `gorm:"size:500" json:"avatarUrl,omitempty"`
	HasSeenOnboarding bool           `gorm:"column:has_seen_onboarding;not null;default:false" json:"hasSeenOnboarding"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}
