package websocket

import (
	"time"

	"github.com/gofiber/websocket/v2"
)

type Client struct {
    conn      *websocket.Conn
    send      chan []byte
    userID    string
    projectID string
}

func (c *Client) WritePump() {
    ticker := time.NewTicker(30 * time.Second)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()

    for {
        select {
        case msg, ok := <-c.send:
            if !ok {
                return
            }
            c.conn.WriteMessage(websocket.TextMessage, msg)

        case <-ticker.C:
            c.conn.WriteMessage(websocket.PingMessage, []byte{})
        }
    }
}
