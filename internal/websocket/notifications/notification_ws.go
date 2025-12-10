package notifications

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func NotificationWS(hub *NotificationHub) fiber.Handler {
    return websocket.New(func(conn *websocket.Conn) {
        userID := conn.Query("user_id")
        if userID == "" {
            conn.Close()
            return
        }

        hub.clients[conn] = userID
        hub.register <- conn

        for {
            if _, _, err := conn.ReadMessage(); err != nil {
                hub.unregister <- conn
                break
            }
        }
    })
}
