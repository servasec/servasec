package models

import "time"

type Team struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"unique;not null;size:100" json:"name"`
	Description string       `gorm:"size:500" json:"description"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
	Members     []TeamMember `gorm:"foreignKey:TeamID" json:"-"`
	MemberCount int64        `gorm:"-" json:"memberCount"`
}
