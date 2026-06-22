package models

import "time"

type Scan struct {
	ID                    uint                 `gorm:"primaryKey" json:"id"`
	ApplicationVersionID  uint                 `gorm:"not null;index" json:"applicationVersionId"`
	ScannerTypeID         uint                 `gorm:"not null;index" json:"scannerTypeId"`
	Status                string               `gorm:"not null;default:pending;size:20" json:"status"`
	StartedAt             *time.Time           `json:"startedAt"`
	CompletedAt           *time.Time           `json:"completedAt"`
	RawResults            *string              `gorm:"type:jsonb" json:"rawResults,omitempty"`
	ApplicationVersion    ApplicationVersion   `gorm:"foreignKey:ApplicationVersionID" json:"applicationVersion"`
	ScannerType           ScannerType          `gorm:"foreignKey:ScannerTypeID" json:"scannerType"`
	Findings              []Finding            `gorm:"foreignKey:ScanID" json:"-"`
	CreatedAt             time.Time            `json:"createdAt"`
	UpdatedAt             time.Time            `json:"updatedAt"`
}
