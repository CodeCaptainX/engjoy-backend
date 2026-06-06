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
	api.Get("/categories", sentenceHandler.showCategories)
	api.Post("/sentences", middleware.AuthMiddleware(db, jwtSecret), sentenceHandler.createSentence)
	api.Get("/sentences", sentenceHandler.show)
	api.Post("/sentences/feed", sentenceHandler.categoryFeed)
	api.Post("/sentences/import/environment", sentenceHandler.importEnvironmentPack)
	api.Post("/sentences/:id/analyze", sentenceHandler.analyzeSentence)
	api.Post("/sentences/:id/review", sentenceHandler.rateSentenceReview)
	api.Delete("/sentences/:id", sentenceHandler.deleteSentence)
	api.Post("/tts", sentenceHandler.generateSpeech)

	// Admin route - strictly authenticated
	admin := api.Group("/admin", middleware.AuthMiddleware(db, jwtSecret))
	admin.Post("/generate-sentences", sentenceHandler.generateSentences)

	// Favorites Routes - require Auth
	favorites := api.Group("/favorites", middleware.AuthMiddleware(db, jwtSecret))
	favorites.Post("/", sentenceHandler.addFavorite)
	favorites.Delete("/:sentenceUuid", sentenceHandler.removeFavorite)
	favorites.Get("/", sentenceHandler.showFavorites)

	// Reactions Routes - require Auth for toggling
	api.Post("/sentences/:sentenceUuid/reaction", middleware.AuthMiddleware(db, jwtSecret), sentenceHandler.toggleReaction)
	api.Get("/sentences/:sentenceUuid/reactions", sentenceHandler.getReactionCount)

	// ... other routes
	api.Post("/sentence", middleware.AuthMiddleware(db, jwtSecret), sentenceHandler.createSentence)
	api.Delete("/sentence/:id", sentenceHandler.deleteSentence)
	api.Get("/sentence", sentenceHandler.show)
	api.Get("/sentence/:id", sentenceHandler.getSentence)

	return sentenceHandler.service
}
