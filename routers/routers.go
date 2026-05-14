package routers

import (
	"errors"

	"sentenceminer/pkg/http/response"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func New() *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			statusCode := fiber.StatusInternalServerError
			message := "internal server error"

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
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, HEAD, PUT, PATCH, POST, DELETE",
	}))

	return app
}
