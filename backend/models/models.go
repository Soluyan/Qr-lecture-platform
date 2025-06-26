// backend/models/models.go
package models

import (
	"sync"
	"time"
)

var (
	Sessions         = make(map[string]Session)
	SessionsLock     sync.Mutex
	SessionQuestions = make(map[string][]Question)
	QuestionsMutex   = &sync.RWMutex{}
)

type Session struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Question struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}
