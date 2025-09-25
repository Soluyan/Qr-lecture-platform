// backend/handlers/broadcast.go
package handlers

import (
	"log"

	"github.com/Soluyan/Qr-lecture-platform/backend/models"
)

/**
 * broadcastQuestions рассылает актуальный список вопросов всем подключенным клиентам в указанной сессии
 * Используется для синхронизации состояния между всеми преподавателями, просматривающими одну сессию
 *
 * @param sessionID - идентификатор сессии, для которой производится рассылка
 * @param questions - актуальный список вопросов для рассылки
 */
func broadcastQuestions(sessionID string, questions []models.Question) {
	hub.RLock()
	clients, ok := hub.clients[sessionID]
	hub.RUnlock()

	if !ok {
		return
	}

	// Итерируемся по всем клиентам в сессии
	for client := range clients {
		go func(c *Client) {
			if err := c.conn.WriteJSON(questions); err != nil {
				log.Println("Broadcast error:", err)
			}
		}(client)
	}
}
