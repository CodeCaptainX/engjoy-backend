package middleware

import (
	"strings"

	response "sentenceminer/pkg/http/response"
	"sentenceminer/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.ApiResponseError(false, "missing authorization header", fiber.StatusUnauthorized, nil))
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.ApiResponseError(false, "invalid authorization header format", fiber.StatusUnauthorized, nil))
		}

		tokenString := parts[1]
		claims, err := utils.VerifyToken(tokenString, jwtSecret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(response.ApiResponseError(false, "invalid or expired token", fiber.StatusUnauthorized, err))
		}

		// Store user ID in locals for subsequent handlers
		c.Locals("userID", claims.UserID)

		return c.Next()
	}
}
