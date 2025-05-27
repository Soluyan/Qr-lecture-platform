package main

import (
    "encoding/json"
    "fmt"
    "github.com/google/uuid"
    "github.com/skip2/go-qrcode"
    "net/http"
    "sync"
    "time"
)

var (
    sessions     = make(map[string]Session)
    sessionsLock sync.Mutex
)

// Session структура для хранения информации о лекции
type Session struct {
    ID        string    `json:"id"`
    ExpiresAt time.Time `json:"expires_at"`
}

// GenerateSessionHandler создает новую сессию и возвращает QR-код
func GenerateSessionHandler(w http.ResponseWriter, r *http.Request) {
    // Генерируем уникальный ID сессии
    sessionID := uuid.New().String()
    
    // Создаем сессию с таймером жизни 80 минут
    expiresAt := time.Now().Add(80 * time.Minute)
    newSession := Session{
        ID:        sessionID,
        ExpiresAt: expiresAt,
    }

    // Сохраняем сессию в памяти
    sessionsLock.Lock()
    sessions[sessionID] = newSession
    sessionsLock.Unlock()

    // Генерируем URL для студентов
    studentURL := fmt.Sprintf("http://yourdomain.com/ask-question?session=%s", sessionID)
    
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

    // Отправляем изображение как ответ
    w.Header().Set("Content-Type", "image/png")
    w.Write(png)
}

// CleanupSessions регулярно очищает просроченные сессии
func CleanupSessions() {
    for {
        time.Sleep(5 * time.Minute) // Проверка каждые 5 минут
        
        sessionsLock.Lock()
        for id, session := range sessions {
            if time.Now().After(session.ExpiresAt) {
                delete(sessions, id)
            }
        }
        sessionsLock.Unlock()
    }
}

func main() {
    go CleanupSessions()
    
    http.HandleFunc("/create-session", GenerateSessionHandler)
    fmt.Println("Server started on :8080")
    http.ListenAndServe(":8080", nil)
}