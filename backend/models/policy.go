package models

import (
	"time"
)

type Policy struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:200;not null" json:"name"`
	Description string    `gorm:"size:1000" json:"description,omitempty"`
	ScopeType   string    `gorm:"size:20;not null" json:"scopeType"`
	ScopeValue  string    `gorm:"size:50;not null" json:"scopeValue"`
	EventTypes  string    `gorm:"type:text;not null" json:"eventTypes"`
	Conditions  string    `gorm:"type:jsonb" json:"conditions,omitempty"`
	Actions     string    `gorm:"type:jsonb" json:"actions,omitempty"`
	IsActive    bool      `gorm:"default:true" json:"isActive"`
	Priority    int       `gorm:"default:0" json:"priority"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type PolicyLog struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	PolicyID       uint      `gorm:"not null;index" json:"policyId"`
	FindingID      uint      `gorm:"not null;index" json:"findingId"`
	EventType      string    `gorm:"size:50" json:"eventType"`
	ConditionsMet  bool      `json:"conditionsMet"`
	ActionType     string    `gorm:"size:50" json:"actionType"`
	ActionResult   string    `gorm:"size:50" json:"actionResult"`
	Detail         string    `gorm:"size:500" json:"detail,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}
