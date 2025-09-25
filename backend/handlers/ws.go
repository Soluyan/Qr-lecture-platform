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

/**
 * upgrader - конфигурация для обновления HTTP соединения до WebSocket
 * CheckOrigin: true разрешает все origin для разработки (в production нужно ограничить)
 */
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

/**
 * Client представляет подключение преподавателя к WebSocket
 * Хранит соединение и идентификатор сессии для маршрутизации сообщений
 */
type Client struct {
	conn      *websocket.Conn
	sessionID string
}

/**
 * Hub - центральный хаб для управления всеми WebSocket подключениями
 * Использует RWMutex для потокобезопасного доступа к мапе клиентов
 * Структура: map[sessionID]map[*Client]bool - группировка клиентов по сессиям
 */
var hub = struct {
	sync.RWMutex
	clients map[string]map[*Client]bool
}{
	clients: make(map[string]map[*Client]bool),
}

/**
 * WsHandler обрабатывает входящие WebSocket соединения от преподавателей
 * Управляет жизненным циклом подключения: регистрация, рассылка, удаление
 */
func WsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("WebSocket connection attempt from:", r.RemoteAddr)

	// Обновляем HTTP соединение до WebSocket протокола
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}

	// Извлекаем ID сессии из query параметров
	sessionID := r.URL.Query().Get("session")
	log.Println("WebSocket session ID:", sessionID)

	// Проверяем наличие sessionID
	if sessionID == "" {
		log.Println("WebSocket: no session ID provided")
		conn.Close()
		return
	}

	// Создаем нового клиента
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

	// Удаление клиента из хаба при выходе из функции
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
		// Читаем сообщение от клиента
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Парсим JSON сообщение
		var request struct {
			Action     string `json:"action"`
			QuestionID string `json:"question_id"`
		}
		if json.Unmarshal(msg, &request) != nil {
			continue
		}

		// Обрабатываем действие удаления
		if request.Action == "delete" {
			deleteQuestion(sessionID, request.QuestionID)
		}
	}
}

/**
 * deleteQuestion удаляет вопрос из хранилища и рассылает обновление
 * Вызывается когда преподаватель удаляет вопрос через интерфейс
 *
 * @param sessionID - идентификатор сессии
 * @param questionID - идентификатор вопроса для удаления
 */
func deleteQuestion(sessionID, questionID string) {
	models.QuestionsMutex.Lock()
	defer models.QuestionsMutex.Unlock()

	// Получаем вопросы для указанной сессии
	questions := models.SessionQuestions[sessionID]

	// Ищем вопрос по ID и удаляем его из слайса
	for i, q := range questions {
		if q.ID == questionID {
			// Удаляем элемент i из слайса
			questions = append(questions[:i], questions[i+1:]...)
			models.SessionQuestions[sessionID] = questions
			log.Println("Question deleted:", questionID, "from session:", sessionID)

			// Рассылаем обновленный список всем подключенным клиентам
			broadcastQuestions(sessionID, questions)
			break
		}
	}
}
