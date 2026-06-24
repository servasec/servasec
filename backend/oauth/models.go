package oauth

import (
	"sync"
	"time"
)

type Client struct {
	ID           string
	Secret       string
	Name         string
	RedirectURIs []string
	Scopes       []string
	CreatedAt    time.Time
}

type AuthCode struct {
	Code            string
	ClientID        string
	UserID          uint
	RedirectURI     string
	ExpiresAt       time.Time
	Challenge       string
	ChallengeMethod string
	Used            bool
}

type Token struct {
	Refresh       string
	ClientID      string
	UserID        uint
	AccessExpiry  time.Time
	RefreshExpiry time.Time
	Revoked       bool
}

type Store struct {
	mu      sync.RWMutex
	clients map[string]*Client
	codes   map[string]*AuthCode
	tokens  map[string]*Token
}

func NewStore() *Store {
	return &Store{
		clients: make(map[string]*Client),
		codes:   make(map[string]*AuthCode),
		tokens:  make(map[string]*Token),
	}
}

func (s *Store) AddClient(c *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[c.ID] = c
}

func (s *Store) GetClient(id string) *Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clients[id]
}

func (s *Store) AddCode(c *AuthCode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.codes[c.Code] = c
}

func (s *Store) GetCode(code string) *AuthCode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.codes[code]
}

func (s *Store) UseCode(code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if c, ok := s.codes[code]; ok {
		c.Used = true
	}
}

func (s *Store) DeleteCode(code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.codes, code)
}

func (s *Store) AddToken(t *Token) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[t.Refresh] = t
}

func (s *Store) GetToken(refresh string) *Token {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tokens[refresh]
}

func (s *Store) RevokeToken(refresh string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.tokens[refresh]; ok {
		t.Revoked = true
	}
}

var DefaultStore = NewStore()
