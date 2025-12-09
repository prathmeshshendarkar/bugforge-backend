package routes

import (
	controller "bugforge-backend/internal/http/controllers/interfaces"

	"github.com/gofiber/fiber/v2"
)

func UserRoutes(router fiber.Router, uc controller.UserController) {
	r := router.Group("/users")

	r.Post("/", uc.CreateUser)
	r.Get("/", uc.GetAllUsers)
	r.Get("/:id", uc.GetUser)
	r.Put("/:id", uc.UpdateUser)
	r.Delete("/:id", uc.DeleteUser)
}
