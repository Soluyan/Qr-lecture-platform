// backend/handlers/ws.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/Soluyan/Qr-lecture-platform/backend/models"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем все origin для разработки
	},
}

// Client представляет подключение преподавателя
type Client struct {
	conn      *websocket.Conn
	sessionID string
}

// Hub управляет подключениями и рассылкой сообщений
var hub = struct {
	sync.RWMutex
	clients map[string]map[*Client]bool // sessionID -> clients
}{
	clients: make(map[string]map[*Client]bool),
}

// WsHandler обрабатывает WebSocket соединения
func WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}

	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		conn.Close()
		return
	}

	client := &Client{
		conn:      conn,
		sessionID: sessionID,
	}

	// Регистрация клиента
	hub.Lock()
	if _, ok := hub.clients[sessionID]; !ok {
		hub.clients[sessionID] = make(map[*Client]bool)
	}
	hub.clients[sessionID][client] = true
	hub.Unlock()

	defer func() {
		hub.Lock()
		delete(hub.clients[sessionID], client)
		if len(hub.clients[sessionID]) == 0 {
			delete(hub.clients, sessionID)
		}
		hub.Unlock()
		conn.Close()
	}()

	// Отправка текущих вопросов при подключении
	models.QuestionsMutex.RLock()
	questions := models.SessionQuestions[sessionID]
	models.QuestionsMutex.RUnlock()
	if err := conn.WriteJSON(questions); err != nil {
		log.Println("Initial send failed:", err)
		return
	}

	// Обработка входящих сообщений
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var request struct {
			Action     string `json:"action"`
			QuestionID string `json:"question_id"`
		}
		if json.Unmarshal(msg, &request) != nil {
			continue
		}

		if request.Action == "delete" {
			deleteQuestion(sessionID, request.QuestionID)
		}
	}
}

// deleteQuestion обрабатывает удаление вопросов
func deleteQuestion(sessionID, questionID string) {
	models.QuestionsMutex.Lock()
	defer models.QuestionsMutex.Unlock()

	questions := models.SessionQuestions[sessionID]
	for i, q := range questions {
		if q.ID == questionID {
			questions = append(questions[:i], questions[i+1:]...)
			models.SessionQuestions[sessionID] = questions
			broadcastQuestions(sessionID, questions)
			break
		}
	}
}
