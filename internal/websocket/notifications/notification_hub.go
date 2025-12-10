package notifications

import (
	"github.com/gofiber/websocket/v2"
)

type NotificationHub struct {
    clients map[*websocket.Conn]string // conn → userID
    register chan *websocket.Conn
    unregister chan *websocket.Conn
    Broadcast chan NotificationWSMessage
}

type NotificationWSMessage struct {
    UserID string      `json:"user_id"`
    Type   string      `json:"type"`
    Data   interface{} `json:"data"`
}

func NewNotificationHub() *NotificationHub {
    return &NotificationHub{
        clients:    make(map[*websocket.Conn]string),
        register:   make(chan *websocket.Conn),
        unregister: make(chan *websocket.Conn),
        Broadcast:  make(chan NotificationWSMessage),
    }
}

func (h *NotificationHub) Run() {
    for {
        select {
        case conn := <-h.register:
            // Handled in handler
            _ = conn

        case conn := <-h.unregister:
            delete(h.clients, conn)
            conn.Close()

        case msg := <-h.Broadcast:
            for conn, userID := range h.clients {
                if userID == msg.UserID { // Send ONLY to target user
                    conn.WriteJSON(msg)
                }
            }
        }
    }
}
