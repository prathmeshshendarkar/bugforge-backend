package websocket

import (
	"bugforge-backend/internal/auth"

	"github.com/gofiber/websocket/v2"
)

func WSHandler(h *CommentHub) func(*websocket.Conn) {
	return func(conn *websocket.Conn) {

		token := conn.Query("token")
		if token == "" {
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"missing token"}`))
			conn.Close()
			return
		}

		// Only userID is needed here
		userID, _, err := auth.VerifyTokenAndGetUserID(token)
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"invalid token"}`))
			conn.Close()
			return
		}


		issueID := conn.Params("issueID")

		room := h.GetRoom(issueID)

		client := &CommentClient{
			conn:    conn,
			send:    make(chan []byte, 256),
			userID:  userID,
			issueID: issueID,
		}

		room.register <- client
		go client.WritePump()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				room.unregister <- client
				return
			}

			// Broadcast to room
			room.Broadcast(msg)
		}
	}
}
