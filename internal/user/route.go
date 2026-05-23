package user

import (
	"sentenceminer/config"
	"sentenceminer/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

func RegisterRoutes(app *fiber.App, db *sqlx.DB) {
	cfg := config.NewConfig()
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	user := app.Group("/api/user")

	favorites := user.Group("/favorites", middleware.AuthMiddleware(cfg.JWTSecret))
	favorites.Post("/add", handler.AddFavorite)
	favorites.Post("/remove", handler.RemoveFavorite)
	favorites.Get("/:userId", handler.GetFavorites)
}
