package interfaces

import "github.com/gofiber/fiber/v2"

type NotificationController interface {
	GetNotifications(c *fiber.Ctx) error
	MarkRead(c *fiber.Ctx) error
	MarkAllRead(c *fiber.Ctx) error
}
