package interfaces

import "github.com/gofiber/fiber/v2"

type ProjectController interface {
    Create(c *fiber.Ctx) error
    GetAll(c *fiber.Ctx) error
    GetByID(c *fiber.Ctx) error
    Update(c *fiber.Ctx) error
    Delete(c *fiber.Ctx) error

    // Optional: if you still want subdomain resolution
    GetBySlug(c *fiber.Ctx) error
}
