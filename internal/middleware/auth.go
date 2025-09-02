package middleware

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/amavis442/til-backend/internal/auth"
	"github.com/amavis442/til-backend/internal/config"
	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(verifier auth.TokenVerifier) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tokenStr string
		if !config.IsProduction() {
			cookie := c.Cookies("access_token")
			slog.Info(fmt.Sprintf("Cookie is '%v'", cookie))
		}

		authHeader := c.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			// Fallback to cookie
			tokenStr = c.Cookies("access_token")
		}
		if !config.IsProduction() {
			slog.Info(fmt.Sprintf("Middleware token string: %v", tokenStr))
		}

		if tokenStr == "" {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		claims, err := verifier.Verify(tokenStr)
		if err != nil || claims["typ"] != "access" {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		userID, err := verifier.ExtractUserID(claims)
		if !config.IsProduction() {
			slog.Info(fmt.Sprintf("User id is: %v", userID))
		}
		if err != nil {
			slog.Info(fmt.Sprintf("Error is: %v", err))
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		c.Locals("userID", userID)
		return c.Next()
	}
}
