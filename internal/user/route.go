package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

func RegisterRoutes(app *fiber.App, db *sqlx.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	api := app.Group("/api")
	api.Post("/login", handler.Login)
}
