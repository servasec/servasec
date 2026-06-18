package models

import "time"

type Scan struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	ApplicationID uint       `gorm:"not null;index" json:"applicationId"`
	ScannerType   string     `gorm:"not null;size:30" json:"scannerType"`
	Status        string     `gorm:"not null;default:pending;size:20" json:"status"`
	StartedAt     *time.Time `json:"startedAt"`
	CompletedAt   *time.Time `json:"completedAt"`
	RawResults    *string    `gorm:"type:jsonb" json:"rawResults,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}
