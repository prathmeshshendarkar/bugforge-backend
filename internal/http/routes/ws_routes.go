package routes

import (
	mw "bugforge-backend/internal/http/middlewares"
	ws "bugforge-backend/internal/websocket"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func RegisterWebSocketRoutes(router fiber.Router, hub *ws.Hub) {

    router.Get("/projects/:projectID",
        mw.JWTProtectedWebSocket(),
        func(c *fiber.Ctx) error {

            if websocket.IsWebSocketUpgrade(c) {
                return c.Next()
            }

            return fiber.ErrUpgradeRequired
        },
        websocket.New(hub.WSHandler),
    )
}
