package models

import (
	"time"

	"gorm.io/gorm"
)

type Application struct {
	ID                uint                 `gorm:"primaryKey" json:"id"`
	Name              string               `gorm:"not null;size:200" json:"name"`
	Description       string               `gorm:"size:1000" json:"description"`
	Slug              string               `gorm:"not null;size:100;uniqueIndex:idx_group_slug" json:"slug"`
	GroupID           uint                 `gorm:"not null;index;uniqueIndex:idx_group_slug" json:"groupId"`
	RepositoryURL     string               `gorm:"size:500" json:"repositoryUrl"`
	ApiToken          string               `gorm:"not null;uniqueIndex;size:64" json:"-"`
	AssetCriticality  string               `gorm:"default:medium;size:10" json:"assetCriticality"`
	Versions          []ApplicationVersion `gorm:"foreignKey:ApplicationID" json:"versions,omitempty"`
	CreatedAt         time.Time            `json:"createdAt"`
	UpdatedAt         time.Time            `json:"updatedAt"`
	DeletedAt         gorm.DeletedAt       `gorm:"index" json:"-"`
}
