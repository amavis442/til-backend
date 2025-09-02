package handler

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/amavis442/til-backend/internal/auth"
	"github.com/amavis442/til-backend/internal/config"
	"github.com/amavis442/til-backend/internal/user"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	userService         user.Service // interface for user validation, etc.
	refreshTokenService auth.Service
	logger              *slog.Logger // Assume Logger interface is defined elsewhere
}

func NewAuthHandler(userSvc user.Service, refreshTokenSvc auth.Service, slogger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		userService:         userSvc,
		refreshTokenService: refreshTokenSvc,
		logger:              slogger,
	}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	isProduction := config.IsProduction()
	if !isProduction {
		h.logger.Info("Login request")
	}

	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Warn(fmt.Sprintf("Failed to parse login request: %v", err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Validate user via service (not hardcoded)
	valid, userID, _ := h.userService.ValidateCredentials(req.Username, req.Password)
	if !valid {
		h.logger.Warn(fmt.Sprintf("Invalid credentials for user: %s", req.Username))
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Login should invalidate the old refresh token if it exists
	// and give new tokens
	err := h.refreshTokenService.DeleteRefreshTokenByUserID(userID)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to remove old refresh token from database for userID %v: %v", userID, err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not generate token"})
	}

	access, refresh, err := auth.GenerateTokens(userID)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to generate tokens for userID %v: %v", userID, err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not generate token"})
	}
	err = h.refreshTokenService.SaveRefreshToken(userID, refresh)
	if err != nil {
		h.logger.Error(err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not generate token"})
	}

	sameSite := "None"
	domain := "localhost"
	if !isProduction {
		sameSite = "Lax"
	}

	cookie := fiber.Cookie{
		Name:     "access_token",
		Value:    access,
		Expires:  time.Now().Add(time.Minute * 15),
		HTTPOnly: true,
		Secure:   isProduction, // Only set to true when using https://
		Domain:   domain,
		SameSite: sameSite,
	}

	c.Cookie(&cookie)

	if !isProduction {
		h.logger.Info(fmt.Sprintf("Access token is: %v", access))
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

	isProduction := config.IsProduction()

	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Warn(fmt.Sprintf("Failed to parse refresh token request: %v", err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	if req.RefreshToken == "" {
		h.logger.Warn("Missing refresh token in request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing refresh token"})
	}

	claims, err := auth.VerifyToken(req.RefreshToken)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("Invalid refresh token: %v", err))
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid refresh token"})
	}

	// Check token type
	if typ, ok := claims["typ"].(string); !ok || typ != "refresh" {
		h.logger.Warn("Invalid token type in refresh token")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token type"})
	}

	userID, err := auth.ExtractUserIDFromClaims(claims)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("Invalid user ID in token: %v", err))
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID in token"})
	}

	refreshToken, err := h.refreshTokenService.FindRefreshTokenByUserID(userID)
	// Check if refresh token is not expired and if so, create a new one. But if it is not expired, check if it is valid.
	if err != nil {
		h.logger.Warn(fmt.Sprintf("No valid refresh token found for userID %v: %v", userID, err))
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No valid refresh token found"})
	}

	// Check if the refresh token send and that stored in the database are the same
	if refreshToken.Token != req.RefreshToken {
		h.logger.Warn(fmt.Sprintf("Refresh token mismatch for userID %v with token: [ %v ]", userID, req.RefreshToken))
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid refresh token found"})
	}

	// Invalidate the used refresh token to prevent replay attacks
	if err := h.refreshTokenService.DeleteRefreshToken(refreshToken.Token); err != nil {
		h.logger.Error(fmt.Sprintf("Could not invalidate refresh token for userID %v: %v", userID, err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not invalidate refresh token"})
	}

	// Generate tokens
	newAccess, newRefresh, err := auth.GenerateTokens(userID)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Could not refresh token for userID %v: %v", userID, err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not refresh token"})
	}

	// Persist new refesh token
	if err := h.refreshTokenService.SaveRefreshToken(userID, newRefresh); err != nil {
		h.logger.Error(fmt.Sprintf("Could not persist new refresh token for userID %v: %v", userID, err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not persist refresh token"})
	}

	sameSite := "None"
	domain := "localhost"
	if !isProduction {
		sameSite = "Lax"
	}

	cookie := fiber.Cookie{
		Name:     "access_token",
		Value:    newAccess,
		Expires:  time.Now().Add(time.Minute * 15),
		HTTPOnly: true,
		Secure:   isProduction,
		Domain:   domain,
		SameSite: sameSite,
	}

	c.Cookie(&cookie)

	if !isProduction {
		h.logger.Info(fmt.Sprintf("New Access token is: %v", newAccess))
	}

	return c.JSON(fiber.Map{
		"access_token":  newAccess,
		"refresh_token": newRefresh,
	})

}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		h.logger.Warn(fmt.Sprintf("Failed to parse register request: %v", err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}
	if req.Username == "" || req.Email == "" || req.Password == "" {
		h.logger.Warn("Missing fields in register request: username=%v, email=%v", req.Username, req.Email)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing fields"})
	}

	err := h.userService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		h.logger.Error("Could not register user %v: %v", req.Username, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not register"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"succes": "User registered"})
}

func (h *AuthHandler) UpdatePassword(c *fiber.Ctx) error {
	var req struct {
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}
	userIDVal := c.Locals("userID")
	userID, ok := userIDVal.(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	if err := h.userService.UpdatePassword(userID, req.Password); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"succes": "User password has been updated"})
}
