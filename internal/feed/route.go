package feed

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

func RegisterRoutes(api fiber.Router, db *sqlx.DB) {
	feedHandler := NewFeedHandler(db)
	api.Get("/learning-feed", feedHandler.getLearningFeed)
}
