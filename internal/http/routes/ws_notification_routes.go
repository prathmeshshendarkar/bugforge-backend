package routes

import (
	"bugforge-backend/internal/websocket/notifications"

	"github.com/gofiber/fiber/v2"
)

func RegisterNotificationWSRoutes(r fiber.Router, hub *notifications.NotificationHub) {
	r.Get("/notifications", notifications.NotificationWS(hub))
}
