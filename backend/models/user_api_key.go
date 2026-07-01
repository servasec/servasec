package models

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/servasec/servasec/backend/utils"
)

type UserApiKey struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"not null;index" json:"userId"`
	Name       string     `gorm:"not null;size:100" json:"name"`
	KeyHash    string     `gorm:"not null;uniqueIndex;size:64" json:"-"`
	KeyPrefix  string     `gorm:"not null;size:8" json:"keyPrefix"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	ExpiresAt  *time.Time `json:"expiresAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}

func GenerateUserApiKey(name string, userID uint) (*UserApiKey, string, error) {
	raw, err := utils.GenerateRandomString(32)
	if err != nil {
		return nil, "", err
	}
	prefix := raw[:8]
	hash := sha256.Sum256([]byte(raw))

	key := &UserApiKey{
		UserID:    userID,
		Name:      name,
		KeyHash:   hex.EncodeToString(hash[:]),
		KeyPrefix: "sc_" + prefix,
	}
	return key, "sc_" + raw, nil
}
