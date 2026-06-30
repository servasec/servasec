package models

import "time"

type Finding struct {
	ID                    uint                 `gorm:"primaryKey" json:"id"`
	ScanID                uint                 `gorm:"not null;index" json:"scanId"`
	ApplicationVersionID  uint                 `gorm:"not null;index" json:"applicationVersionId"`
	ScannerTypeID         uint                 `gorm:"not null;index" json:"scannerTypeId"`
	RuleID                string               `gorm:"size:200;index" json:"ruleId"`
	Title                 string               `gorm:"size:500" json:"title"`
	Severity              string               `gorm:"not null;size:10;index" json:"severity"`
	Description           string               `gorm:"type:text" json:"description"`
	FilePath              string               `gorm:"size:1000" json:"filePath"`
	LineStart             *int                 `json:"lineStart"`
	LineEnd               *int                 `json:"lineEnd"`
	CWEID                 string               `gorm:"size:200" json:"cweId"`
	Remediation           string               `gorm:"type:text" json:"remediation"`
	DedupeHash            string               `gorm:"size:64;index" json:"dedupeHash,omitempty"`
	Status                string               `gorm:"not null;default:open;size:20;index" json:"status"`
	AssignedTo            *uint                `gorm:"index" json:"assignedTo"`
	AssignedToUser        *User                `gorm:"foreignKey:AssignedTo" json:"assignedToUser,omitempty"`
	DueDate               *time.Time           `json:"dueDate"`
	ReviewedBy            *uint                `gorm:"index" json:"reviewedBy"`
	ReviewedByUser        *User                `gorm:"foreignKey:ReviewedBy" json:"reviewedByUser,omitempty"`
	RiskScore             *float64             `gorm:"index" json:"riskScore"`
	EPSSScore             *float64             `json:"epssScore"`
	FixedAt               *time.Time           `json:"fixedAt"`
	Scan                  Scan                 `gorm:"foreignKey:ScanID" json:"-"`
	ApplicationVersion    ApplicationVersion   `gorm:"foreignKey:ApplicationVersionID" json:"applicationVersion"`
	ScannerType           ScannerType          `gorm:"foreignKey:ScannerTypeID" json:"scannerType"`
	CreatedAt             time.Time            `json:"createdAt"`
	UpdatedAt             time.Time            `json:"updatedAt"`
}
