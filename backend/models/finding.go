package models

import "time"

type Finding struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ScanID        uint      `gorm:"not null;index" json:"scanId"`
	ApplicationID uint      `gorm:"not null;index" json:"applicationId"`
	ScannerType   string    `gorm:"not null;size:30" json:"scannerType"`
	RuleID        string    `gorm:"size:200;index" json:"ruleId"`
	Title         string    `gorm:"size:500" json:"title"`
	Severity      string    `gorm:"not null;size:10;index" json:"severity"`
	Description   string    `gorm:"type:text" json:"description"`
	FilePath      string    `gorm:"size:1000" json:"filePath"`
	LineStart     *int      `json:"lineStart"`
	LineEnd       *int      `json:"lineEnd"`
	CWEID         string    `gorm:"size:20" json:"cweId"`
	Remediation   string    `gorm:"type:text" json:"remediation"`
	Status        string    `gorm:"not null;default:open;size:20;index" json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
