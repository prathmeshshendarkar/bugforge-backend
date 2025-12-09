package interfaces

import "github.com/gofiber/fiber/v2"

type ProjectLabelController interface {
	List(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
}
