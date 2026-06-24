package models

import "time"

type AuditLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"userId" gorm:"index"`
	Username  string    `json:"username" gorm:"size:100"`
	Action    string    `json:"action" gorm:"size:10"`     // POST, PUT, PATCH, DELETE
	Resource  string    `json:"resource" gorm:"size:500"`  // path
	Status    int       `json:"status"`                     // HTTP status code
	Details   string    `json:"details,omitempty" gorm:"type:text"`
	IP        string    `json:"ip" gorm:"size:50"`
	CreatedAt time.Time `json:"createdAt"`
}
