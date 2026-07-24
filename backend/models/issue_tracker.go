package models

import (
	"time"

	"gorm.io/gorm"
)

type IssueTracker struct {
	ID                      uint           `gorm:"primaryKey" json:"id"`
	ApplicationID           uint           `gorm:"not null;uniqueIndex" json:"applicationId"`
	Application             Application    `gorm:"foreignKey:ApplicationID" json:"-"`
	Provider                string         `gorm:"not null;size:20" json:"provider"`
	AuthType                string         `gorm:"not null;default:pat;size:20" json:"authType"`
	EncryptedToken          string         `gorm:"size:500" json:"-"`
	EncryptedTokenIV        string         `gorm:"size:44" json:"-"`
	GitHubAppID             *int64         `gorm:"column:github_app_id" json:"-"`
	GitHubInstallationID    *int64         `gorm:"column:github_installation_id" json:"-"`
	EncryptedGitHubAppKey   string         `gorm:"column:encrypted_github_app_key;size:6000" json:"-"`
	EncryptedGitHubAppKeyIV string         `gorm:"column:encrypted_github_app_key_iv;size:44" json:"-"`
	TokenExpiresAt          *time.Time     `json:"-"`
	SeverityThreshold       string         `gorm:"not null;default:high;size:10" json:"severityThreshold"`
	IsActive                bool           `gorm:"default:true" json:"isActive"`
	CreatedAt               time.Time      `json:"createdAt"`
	UpdatedAt               time.Time      `json:"updatedAt"`
	DeletedAt               gorm.DeletedAt `gorm:"index" json:"-"`
}
