package routes

import (
	controller "bugforge-backend/internal/http/controllers/interfaces"

	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(public fiber.Router, ac controller.AuthController) {
    // PUBLIC
    public.Post("/login", ac.Login)
	public.Post("/accept-invite", ac.AcceptInvite)
}

func AuthProtectedRoutes(protected fiber.Router, ac controller.AuthController) {
    // PROTECTED
    protected.Get("/me", ac.Me)
}
