package routes

import (
	ctrl "bugforge-backend/internal/http/controllers/interfaces"

	"github.com/gofiber/fiber/v2"
)

func NotificationRoutes(router fiber.Router, nc ctrl.NotificationController) {
	r := router.Group("/notifications")

	r.Get("/", nc.GetNotifications)
	r.Patch("/:id/read", nc.MarkRead)
	r.Patch("/read-all", nc.MarkAllRead)
}
	