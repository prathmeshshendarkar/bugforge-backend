package websocket

import (
	"encoding/json"

	"github.com/gofiber/websocket/v2"
)

func (h *Hub) WSHandler(conn *websocket.Conn) {
	projectID := conn.Params("projectID")

	userID, ok := conn.Locals("user_id").(string)
	if !ok {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"unauthorized"}`))
		conn.Close()
		return
	}

	room := h.GetRoom(projectID)

	client := &Client{
		conn:      conn,
		send:      make(chan []byte, 256),
		userID:    userID,
		projectID: projectID,
	}

	room.register <- client
	go client.WritePump()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			room.unregister <- client
			break
		}

		var evt IncomingEvent
		if err := json.Unmarshal(msg, &evt); err != nil {
			continue
		}

		switch evt.Type {

		case "move_card":
			// call your logic
			out, err := HandleMoveCardIntent(evt, userID)
			if err == nil {
				b, _ := json.Marshal(out)
				room.Broadcast(b)
			}

		case "column_created":
			b, _ := json.Marshal(evt)
			room.Broadcast(b)

		case "card_created":
			b, _ := json.Marshal(evt)
			room.Broadcast(b)
		}
	}
}
