// backend/handlers/broadcast.go
package main

import "log"

// broadcastQuestions рассылает вопросы всем клиентам в сессии
func broadcastQuestions(sessionID string, questions []Question) {
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
