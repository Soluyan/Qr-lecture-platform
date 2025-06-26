// backend/models/session.go
package main

import (
	"sync"
	"time"
)

var (
	sessions     = make(map[string]Session)
	sessionsLock sync.Mutex
)

type Session struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
}
