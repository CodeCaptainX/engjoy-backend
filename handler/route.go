package handler

import (
	"sentenceminer/internal/sentences"
	"sentenceminer/internal/user"
	"sentenceminer/pkg/http/response"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type RouteDependencies struct {
	DB *sqlx.DB
}

func RegisterRoutes(app *fiber.App, deps RouteDependencies) {
	api := app.Group("/api")

	api.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(response.NewResponse("health ok", fiber.StatusOK, fiber.Map{"status": "ok"}))
	})

	sentences.RegisterRoutes(app, deps.DB)
	user.RegisterRoutes(app, deps.DB)
}
