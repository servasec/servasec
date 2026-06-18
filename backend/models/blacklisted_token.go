package models

import "time"

type BlacklistedToken struct {
	TokenHash string    `gorm:"primaryKey;size:64"`
	CreatedAt time.Time
}
