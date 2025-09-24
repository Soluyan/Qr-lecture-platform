// backend/models/models.go
package models

import (
	"sync"
	"time"
)

var (
	Sessions         = make(map[string]Session)
	SessionsLock     = &sync.RWMutex{}
	SessionQuestions = make(map[string][]Question)
	QuestionsMutex   = &sync.RWMutex{}
)

type SessionSettings struct {
	AllowAnonymous bool `json:"allowAnonymous"`
}

type Session struct {
	ID        string          `json:"id"`
	ExpiresAt time.Time       `json:"expiresAt"`
	Settings  SessionSettings `json:"settings"`
}

type Question struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	Author    string `json:"author"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}
