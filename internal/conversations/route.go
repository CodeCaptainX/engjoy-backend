package conversations

import (
	"sentenceminer/internal/conversations/service"
	"sentenceminer/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// RegisterRoutes now accepts the Fiber App to properly manage top-level route grouping
func RegisterRoutes(api fiber.Router, db *sqlx.DB, jwtSecret string) *service.ConversationService {
	conversationHandler := NewConversationHandler(db)
	
	// Middleware
	auth := middleware.AuthMiddleware(db, jwtSecret)
	optionalAuth := middleware.OptionalAuthMiddleware(db, jwtSecret)

	// Public routes
	api.Get("/conversations", optionalAuth, conversationHandler.showUserConversations)
	api.Get("/conversations/:id", optionalAuth, conversationHandler.getUserConversation)
	api.Get("/conversations/:conversationId/messages", optionalAuth, conversationHandler.listUserConversationMessages)
	api.Get("/conversations/:conversationId/messages/:messageId", optionalAuth, conversationHandler.getUserConversationMessage)

	// Authenticated routes
	api.Post("/conversations", auth, conversationHandler.createUserConversation)
	api.Patch("/conversations/:id", auth, conversationHandler.updateUserConversation)
	api.Delete("/conversations/:id", auth, conversationHandler.deleteUserConversation)

	api.Post("/conversations/:conversationId/messages", auth, conversationHandler.addMessageToUserConversation)
	api.Post("/conversations/:conversationId/messages/batch", auth, conversationHandler.addMultipleMessagesToUserConversation)
	api.Patch("/conversations/:conversationId/messages/:messageId", auth, conversationHandler.updateUserConversationMessage)
	api.Delete("/conversations/:conversationId/messages/:messageId", auth, conversationHandler.deleteUserConversationMessage)

	return conversationHandler.service
}
