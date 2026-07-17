package models

import (
	"time"

	"gorm.io/gorm"
)

type Group struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"not null;size:100" json:"name"`
	Description string         `gorm:"size:500" json:"description"`
	Path        string         `gorm:"unique;not null;size:100" json:"path"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
