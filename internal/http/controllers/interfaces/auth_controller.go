package interfaces

import "github.com/gofiber/fiber/v2"

type AuthController interface {
	Login(c *fiber.Ctx) error
	Me(c *fiber.Ctx) error
	AcceptInvite(c *fiber.Ctx) error
}
