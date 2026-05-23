package auth

import (
	"sentenceminer/config"
	"sentenceminer/internal/user"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

func RegisterRoutes(app *fiber.App, db *sqlx.DB) {
	cfg := config.NewConfig()
	userRepo := user.NewRepository(db)
	service := NewService(cfg, userRepo)
	handler := NewHandler(service)

	auth := app.Group("/api/auth")
	auth.Post("/login", handler.Login)
	auth.Get("/google/login", handler.GoogleLogin)
	auth.Get("/google/callback", handler.GoogleCallback)
}
