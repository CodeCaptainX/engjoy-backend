package middleware

import (
	"strings"

	"sentenceminer/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// OptionalAuthMiddleware attempts to verify a JWT token.
// If valid, it stores userID in Locals. If missing or invalid, it proceeds without error.
func OptionalAuthMiddleware(db *sqlx.DB, jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next() // No token, proceed as anonymous
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Next() // Invalid format, treat as anonymous
		}

		tokenString := parts[1]
		claims, err := utils.VerifyToken(tokenString, jwtSecret)
		if err != nil {
			return c.Next() // Invalid token, proceed as anonymous
		}

		// Check if user still exists in DB
		var exists bool
		err = db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM tbl_users WHERE id = $1 AND deleted_at IS NULL)", claims.UserID)
		if err != nil || !exists {
			return c.Next() // User not found, proceed as anonymous
		}

		// Store user ID in locals for subsequent handlers
		c.Locals("userID", claims.UserID)

		return c.Next()
	}
}
