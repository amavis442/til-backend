package handler

import (
	"github.com/amavis442/til-backend/internal/auth"
	"github.com/amavis442/til-backend/internal/user"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	userService user.Service // interface for user validation, etc.
}

func NewAuthHandler(userSvc user.Service) *AuthHandler {
	return &AuthHandler{
		userService: userSvc,
	}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Validate user via service (not hardcoded)
	valid, userID, _ := h.userService.ValidateCredentials(req.Username, req.Password)
	if !valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	access, refresh, err := auth.GenerateTokens(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not generate token"})
	}

	return c.JSON(fiber.Map{
		"access_token":  access,
		"refresh_token": refresh,
	})
}

func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token"`
	}

	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	if req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing refresh token"})
	}

	claims, err := auth.VerifyToken(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid refresh token"})
	}

	// Check token type
	if typ, ok := claims["typ"].(string); !ok || typ != "refresh" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token type"})
	}

	userID, err := auth.ExtractUserIDFromClaims(claims)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID in token"})
	}

	newAccess, newRefresh, err := auth.GenerateTokens(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not refresh token"})
	}

	return c.JSON(fiber.Map{
		"access_token":  newAccess,
		"refresh_token": newRefresh,
	})
}
