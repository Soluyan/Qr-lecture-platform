// backend/main.go
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

/**
 * GenerateSessionHandler создает новую сессию лекции
 * Генерирует уникальный ID сессии, QR-код и настройки по умолчанию
 * Возвращает JSON с sessionId и QR-кодом в base64
 */
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
	expiresAt := time.Now().Add(90 * time.Minute)
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

	// Генерируем URL для студентов
	baseURL := os.Getenv("RAILWAY_STATIC_URL")
	if baseURL == "" {
		baseURL = "https://qr-lecture-platform-production.up.railway.app"
	}
	studentURL := fmt.Sprintf("%s/student?session=%s", baseURL, sessionID)

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

/**
 * UpdateSessionSettingsHandler обновляет настройки существующей сессии
 * Позволяет изменять разрешение анонимных вопросов
 */
func UpdateSessionSettingsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем ID сессии из query
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	// Парсим JSON
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

	// Возвращаем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}

/**
 * GetSessionSettingsHandler возвращает текущие настройки сессии
 * Используется студенческой частью для проверки разрешения анонимных вопросов
 */
func GetSessionSettingsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Извлекаем ID сессии из query
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	// Читаем данные сессии
	models.SessionsLock.RLock()
	session, exists := models.Sessions[sessionID]
	models.SessionsLock.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Возвращаем настройки сессии в JSON формате
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session.Settings)
}

/**
 * CleanupSessions периодически очищает просроченные сессии и вопросы
 * Запускается в отдельной goroutine и работает каждые 5 минут
 */
func CleanupSessions() {
	for {
		time.Sleep(5 * time.Minute)

		// Блокируем доступ к данным для безопасной очистки
		models.SessionsLock.Lock()
		models.QuestionsMutex.Lock()

		// Проходим по всем сессиям и удаляем просроченные
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

/**
 * enableCORS middleware добавляет CORS заголовки к HTTP обработчикам
 * Позволяет фронтенду на другом порту взаимодействовать с бэкендом
 */
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Разрешаем запросы с любого origin в продакшене
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

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
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// Создаем HTTP сервер с настройками
	server := &http.Server{
		Addr: ":" + os.Getenv("PORT"),
	}

	go CleanupSessions()

	// Статические файлы фронтенда
	http.Handle("/", http.FileServer(http.Dir("./public")))

	// WebSocket endpoint для реального времени
	http.HandleFunc("/ws", handlers.WsHandler)

	// REST API endpoints с CORS поддержкой
	http.HandleFunc("/create-session", enableCORS(GenerateSessionHandler))
	http.HandleFunc("/ask", enableCORS(handlers.AskQuestionHandler))
	http.HandleFunc("/session/settings", enableCORS(UpdateSessionSettingsHandler))
	http.HandleFunc("/session/settings/get", enableCORS(GetSessionSettingsHandler))

	http.HandleFunc("/student", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/index.html")
	})

	log.Printf("Server starting on port %s...", os.Getenv("PORT"))

	// Запускаем сервер в отдельной горутине
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Создаем канал для получения сигналов ОС
	quit := make(chan os.Signal, 1)

	// Регистрируем обработчики для сигналов завершения
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// Ожидаем завершение
	<-quit

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Останавливаем сервер с таймаутом
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
}
