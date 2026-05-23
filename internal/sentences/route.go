package sentences

import (
	"sentenceminer/internal/sentences/service"
	"sentenceminer/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

func RegisterRoutes(app *fiber.App, db *sqlx.DB, jwtSecret string) *service.SentenceService {
	api := app.Group("/api")

	sentenceHandler := NewSentenceHandler(db)
	api.Post("/sentences", sentenceHandler.createSentence)
	api.Get("/sentences", sentenceHandler.show)
	api.Post("/sentences/feed", sentenceHandler.categoryFeed)
	api.Post("/sentences/import/environment", sentenceHandler.importEnvironmentPack)
	api.Post("/sentences/:id/analyze", sentenceHandler.analyzeSentence)
	api.Post("/sentences/:id/review", sentenceHandler.rateSentenceReview)
	api.Delete("/sentences/:id", sentenceHandler.deleteSentence)
	api.Post("/tts", sentenceHandler.generateSpeech)

	admin := api.Group("/admin", middleware.AuthMiddleware(jwtSecret))
	admin.Post("/generate-sentences", sentenceHandler.generateSentences)

	api.Post("/sentence", sentenceHandler.createSentence)
	api.Get("/sentence", sentenceHandler.show)
	api.Get("/sentence/:id", sentenceHandler.getSentence)

	return sentenceHandler.service
}
