// backend/handlers/broadcast.go
package handlers

import (
	"log"

	"github.com/Soluyan/Qr-lecture-platform/backend/models"
)

// broadcastQuestions рассылает вопросы всем клиентам в сессии
func broadcastQuestions(sessionID string, questions []models.Question) {
	hub.RLock()
	clients, ok := hub.clients[sessionID]
	hub.RUnlock()

	if !ok {
		return
	}

	for client := range clients {
		go func(c *Client) {
			if err := c.conn.WriteJSON(questions); err != nil {
				log.Println("Broadcast error:", err)
			}
		}(client)
	}
}
