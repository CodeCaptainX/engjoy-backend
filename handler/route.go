package handler

import (
	"sentenceminer/config"
	"sentenceminer/internal/auth"
	"sentenceminer/internal/conversations"
	"sentenceminer/internal/feed"
	"sentenceminer/internal/sentences"
	"sentenceminer/internal/user"
	response "sentenceminer/pkg/http/response"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type RouteDependencies struct {
	DB     *sqlx.DB
	Config config.AppConfig
}

func RegisterRoutes(app *fiber.App, deps RouteDependencies) {
	api := app.Group("/api")

	api.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(response.NewResponse(c, "health ok", fiber.StatusOK, fiber.Map{"status": "ok"}))
	})

	sentenceSvc := sentences.RegisterRoutes(app, deps.DB, deps.Config.JWTSecret)
	sentenceSvc.StartDailyGenerationWorker()

	user.RegisterRoutes(app, deps.DB)
	auth.RegisterRoutes(app, deps.DB)
	conversations.RegisterRoutes(api, deps.DB, deps.Config.JWTSecret)
	feed.RegisterRoutes(api, deps.DB)
}
