package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Setup(app *fiber.App, db *pgxpool.Pool) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"Status": "Ok"});
	})
}