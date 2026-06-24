package mcp

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type Session struct {
	ID     string
	Ch     chan []byte
	Done   chan struct{}
	UserID uint
}

var (
	sessionsMu sync.RWMutex
	sessions   = map[string]*Session{}
)

func generateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func NewSession(userID uint) *Session {
	s := &Session{
		ID:     generateSessionID(),
		Ch:     make(chan []byte, 100),
		Done:   make(chan struct{}),
		UserID: userID,
	}
	sessionsMu.Lock()
	sessions[s.ID] = s
	sessionsMu.Unlock()
	return s
}

func GetSession(id string) *Session {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()
	return sessions[id]
}

func DeleteSession(id string) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	if s, ok := sessions[id]; ok {
		close(s.Done)
		delete(sessions, id)
	}
}
