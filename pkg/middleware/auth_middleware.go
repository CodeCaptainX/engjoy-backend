package middleware

import (
	"strings"

	response "sentenceminer/pkg/http/response"
	"sentenceminer/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

func AuthMiddleware(db *sqlx.DB, jwtSecret string, publicPaths ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		for _, publicPath := range publicPaths {
			if strings.HasPrefix(path, publicPath) {
				return c.Next()
			}
		}

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

		// Check if user still exists in DB
		var exists bool
		err = db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM tbl_users WHERE id = $1 AND deleted_at IS NULL)", claims.UserID)
		if err != nil || !exists {
			return c.Status(fiber.StatusUnauthorized).JSON(response.ApiResponseError(false, "user account not found or disabled", fiber.StatusUnauthorized, nil))
		}

		// Store user ID in locals for subsequent handlers
		c.Locals("userID", claims.UserID)

		return c.Next()
	}
}
