// backend/handlers/question.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Soluyan/Qr-lecture-platform/backend/models"
	"github.com/google/uuid"
)

/**
 * AskQuestionHandler обрабатывает HTTP POST запросы на создание новых вопросов от студентов
 * Выполняет валидацию, сохраняет вопрос и рассылает обновление через WebSocket
 *
 * @param w - HTTP ResponseWriter для формирования ответа
 * @param r - HTTP Request с данными вопроса
 */
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

	// Декодируем JSON тело запроса
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Проверяем обязательные поля
	if req.Author == "" || req.Text == "" {
		http.Error(w, "Author and text are required", http.StatusBadRequest)
		return
	}

	// Проверяем максимальную длину вопроса
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
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	// Блокируем мьютекс для безопасной работы с разделяемыми данными
	models.QuestionsMutex.Lock()
	// Добавляем вопрос в слайс вопросов данной сессии
	models.SessionQuestions[sessionID] = append(models.SessionQuestions[sessionID], question)
	// Получаем актуальный список вопросов для рассылки
	questions := models.SessionQuestions[sessionID]
	models.QuestionsMutex.Unlock()

	log.Println("Question added:", question.ID, "from session:", sessionID)
	// Рассылаем обновленный список вопросов
	broadcastQuestions(sessionID, questions)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(question)
}
