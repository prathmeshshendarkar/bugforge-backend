package interfaces

import "github.com/gofiber/fiber/v2"

type UserController interface {
	CreateUser(c *fiber.Ctx) error
	GetUser(c *fiber.Ctx) error
	GetAllUsers(c *fiber.Ctx) error
	UpdateUser(c *fiber.Ctx) error
	DeleteUser(c *fiber.Ctx) error
}
