// GenerateSessionHandler creates a new lecture session and returns QR code
// @Summary Create new session
// @Description Generates new lecture session with 80min lifetime and returns QR code
// @Produce png
// @Success 200 {file} binary "QR code image"
// @Router /create-session [get]

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Soluyan/Qr-lecture-platform/backend/handlers"
	"github.com/Soluyan/Qr-lecture-platform/backend/models"
	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
)

// GenerateSessionHandler создает новую сессию и возвращает JSON с sessionId и QR-кодом
func GenerateSessionHandler(w http.ResponseWriter, r *http.Request) {
	// Добавляем CORS заголовки
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Генерируем уникальный ID сессии
	sessionID := uuid.New().String()

	// Создаем сессию с таймером жизни 80 минут
	expiresAt := time.Now().Add(80 * time.Minute)
	newSession := models.Session{
		ID:        sessionID,
		ExpiresAt: expiresAt,
		Settings: models.SessionSettings{
			AllowAnonymous: true, // По умолчанию разрешены анонимные вопросы
		},
	}

	// Сохраняем сессию в памяти
	models.SessionsLock.Lock()
	models.Sessions[sessionID] = newSession
	models.SessionsLock.Unlock()

	// Генерируем URL для студентов (используем localhost для разработки)
	studentURL := fmt.Sprintf("http://localhost:5173/ask?session=%s", sessionID)

	// Создаем QR-код
	qr, err := qrcode.New(studentURL, qrcode.Medium)
	if err != nil {
		http.Error(w, "Error generating QR code", http.StatusInternalServerError)
		return
	}

	// Конвертируем в PNG изображение
	png, err := qr.PNG(256)
	if err != nil {
		http.Error(w, "Error generating QR image", http.StatusInternalServerError)
		return
	}

	// Кодируем QR-код в base64
	qrBase64 := base64.StdEncoding.EncodeToString(png)

	// Отправляем JSON ответ с sessionId и QR-кодом
	response := map[string]string{
		"sessionId": sessionID,
		"qrCode":    qrBase64,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func UpdateSessionSettingsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	var settings struct {
		AllowAnonymous bool `json:"allowAnonymous"`
	}

	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Обновляем настройки сессии
	models.SessionsLock.Lock()
	if session, exists := models.Sessions[sessionID]; exists {
		session.Settings.AllowAnonymous = settings.AllowAnonymous
		models.Sessions[sessionID] = session
	}
	models.SessionsLock.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}

func GetSessionSettingsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	models.SessionsLock.RLock()
	session, exists := models.Sessions[sessionID]
	models.SessionsLock.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session.Settings)
}

// CleanupSessions регулярно очищает просроченные сессии
func CleanupSessions() {
	for {
		time.Sleep(5 * time.Minute)
		models.SessionsLock.Lock()
		models.QuestionsMutex.Lock()
		for id, session := range models.Sessions {
			if time.Now().After(session.ExpiresAt) {
				delete(models.Sessions, id)
				delete(models.SessionQuestions, id)
			}
		}
		models.QuestionsMutex.Unlock()
		models.SessionsLock.Unlock()
	}
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	server := &http.Server{
		Addr: ":8080",
	}

	go CleanupSessions()

	http.Handle("/", http.FileServer(http.Dir("../frontend/public")))
	http.HandleFunc("/ws", handlers.WsHandler)
	http.HandleFunc("/create-session", enableCORS(GenerateSessionHandler))
	http.HandleFunc("/ask", enableCORS(handlers.AskQuestionHandler))
	http.HandleFunc("/session/settings", enableCORS(UpdateSessionSettingsHandler))
	http.HandleFunc("/session/settings/get", enableCORS(GetSessionSettingsHandler))

	log.Println("Server starting on :8080...")

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
}
