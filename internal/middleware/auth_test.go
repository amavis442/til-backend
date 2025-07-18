package middleware_test

import (
	"net/http/httptest"
	"testing"

	"github.com/amavis442/til-backend/internal/auth"
	"github.com/amavis442/til-backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func TestAuthMiddleware(t *testing.T) {
	// Setup a test Fiber app with your middleware and a dummy protected route
	app := fiber.New()
	app.Use(middleware.AuthMiddleware)
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Generate a valid token for testing
	validToken, _, err := auth.GenerateTokens(1)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "Valid token",
			token:          validToken,
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "Missing token",
			token:          "",
			expectedStatus: fiber.StatusUnauthorized,
		},
		{
			name:           "Invalid token",
			token:          "some.invalid.token",
			expectedStatus: fiber.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/protected", nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}
		})
	}
}
