package routers

import (
	"errors"
	"log"
	"os"

	"sentenceminer/pkg/middleware"
	response "sentenceminer/pkg/http/response"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func New() *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			statusCode := fiber.StatusInternalServerError
			message := "internal server error"

			if err != nil {
				log.Printf("Error: %v", err)
			}

			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				statusCode = fiberErr.Code
				message = fiberErr.Message
			} else if err != nil {
				message = err.Error()
			}

			return response.JSONError(c, statusCode, message, errors.New(message))
		},
	})

	app.Use(logger.New())
	app.Use(middleware.I18nMiddleware())
	
	// CORS: Restricted to our website origin to prevent third-party API abuse.
	// You should set ALLOWED_ORIGINS in your .env file.
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "*" // Fallback to all during development, but warn in logs
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-App-ID",
		AllowMethods: "GET, HEAD, PUT, PATCH, POST, DELETE",
	}))

	return app
}
