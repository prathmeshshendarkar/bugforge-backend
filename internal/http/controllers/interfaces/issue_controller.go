package interfaces

import "github.com/gofiber/fiber/v2"

type IssueController interface {
	Create(c *fiber.Ctx) error
	ListAll(c *fiber.Ctx) error
	Get(c *fiber.Ctx) error
	ListByProject(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
	
	ListActivity(c *fiber.Ctx) error

	UpdateDueDate(ctx *fiber.Ctx) error
}
