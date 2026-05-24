package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func I18nMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		lang := c.Get("Accept-Language")
		if strings.HasPrefix(lang, "km") {
			c.Locals("lang", "km")
		} else {
			c.Locals("lang", "en")
		}
		return c.Next()
	}
}
