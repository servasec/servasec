package models

import (
	"time"

	"gorm.io/gorm"
)

type ApplicationVersion struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	ApplicationID uint           `gorm:"not null;uniqueIndex:idx_app_version_name" json:"applicationId"`
	Name          string         `gorm:"not null;size:100;uniqueIndex:idx_app_version_name" json:"name"`
	Branch        string         `gorm:"size:200" json:"branch"`
	Tag           string         `gorm:"size:100" json:"tag"`
	IsDefault     bool           `gorm:"default:false" json:"isDefault"`
	Application   Application    `gorm:"foreignKey:ApplicationID" json:"-"`
	Scans         []Scan         `gorm:"foreignKey:ApplicationVersionID" json:"-"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
