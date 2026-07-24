package models

import "time"

type IssueTrackerIssue struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	IssueTrackerID   uint      `gorm:"not null;uniqueIndex:idx_tracker_finding" json:"issueTrackerId"`
	FindingID        uint      `gorm:"not null;uniqueIndex:idx_tracker_finding" json:"findingId"`
	Finding          Finding   `gorm:"foreignKey:FindingID" json:"-"`
	ExternalIssueID  string    `gorm:"not null;size:100" json:"externalIssueId"`
	ExternalIssueURL string    `gorm:"not null;size:500" json:"externalIssueUrl"`
	Status           string    `gorm:"not null;default:open;size:20" json:"status"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}
