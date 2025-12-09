package interfaces

import "github.com/gofiber/fiber/v2"

type ClientController interface {
	GetByID(c *fiber.Ctx) error
}