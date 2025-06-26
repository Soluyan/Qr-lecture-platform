// backend/models/question.go
package models

import (
	"sync"
	"time"
)

var (
	sessionQuestions      = make(map[string][]Question)
	sessionQuestionsMutex = &sync.RWMutex{}
)

// Question структура для хранения вопроса
type Question struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}
