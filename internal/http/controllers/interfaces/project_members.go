package interfaces

import "github.com/gofiber/fiber/v2"

type ProjectMemberController interface {
	List(c *fiber.Ctx) error 
	Add(c *fiber.Ctx) error
	Remove(c *fiber.Ctx) error
	Invite(c *fiber.Ctx) error
}
