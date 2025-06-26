// backend/handlers/question.go
package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Soluyan/Qr-lecture-platform/backend/models"
	"github.com/google/uuid"
)

// AskQuestionHandler обрабатывает вопросы от студентов
func AskQuestionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем sessionID из URL
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	// Проверяем существование сессии
	models.SessionsLock.Lock()
	_, exists := models.Sessions[sessionID]
	models.SessionsLock.Unlock()
	if !exists {
		http.Error(w, "Session not found or expired", http.StatusNotFound)
		return
	}

	// Парсим JSON тело запроса
	var req struct {
		Author string `json:"author"`
		Text   string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Author == "" || req.Text == "" {
		http.Error(w, "Author and text are required", http.StatusBadRequest)
		return
	}
	if len(req.Text) > 500 {
		http.Error(w, "Question is too long (max 500 chars)", http.StatusBadRequest)
		return
	}

	// Создаем новый вопрос
	question := models.Question{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Author:    req.Author,
		Text:      req.Text,
		CreatedAt: time.Now(),
	}

	// Добавляем вопрос в хранилище
	models.QuestionsMutex.Lock()
	models.SessionQuestions[sessionID] = append(models.SessionQuestions[sessionID], question)
	questions := models.SessionQuestions[sessionID]
	models.QuestionsMutex.Unlock()

	// Рассылаем обновленный список вопросов
	broadcastQuestions(sessionID, questions)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(question)
}
