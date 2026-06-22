package models

import "time"

type ScannerType struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"unique;not null;size:30" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	Parser      string    `gorm:"size:100" json:"parser"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
