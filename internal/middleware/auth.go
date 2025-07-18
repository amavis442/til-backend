package middleware

import (
	"strings"

	"github.com/amavis442/til-backend/internal/auth"
	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	claims, err := auth.VerifyToken(tokenStr)
	if err != nil || claims["typ"] != "access" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	userID, err := auth.ExtractUserIDFromClaims(claims)
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	c.Locals("userID", userID)
	return c.Next()
}
