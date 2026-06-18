package models

import "time"

type TeamMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TeamID    uint      `gorm:"not null;uniqueIndex:idx_team_user" json:"teamId"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_team_user" json:"userId"`
	Role      string    `gorm:"not null;default:member;size:20" json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}
