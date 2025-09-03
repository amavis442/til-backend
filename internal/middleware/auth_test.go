package middleware_test

import (
	"errors"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/amavis442/til-backend/internal/config"
	"github.com/amavis442/til-backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type mockTokenVerifier struct{}

func (m *mockTokenVerifier) Verify(tokenStr string) (jwt.MapClaims, error) {
	if tokenStr == "some.invalid.token" {
		return nil, errors.New("invalid token")
	}
	return jwt.MapClaims{
		"sub": "1",
		"typ": "access",
		"exp": float64(time.Now().Add(15 * time.Minute).Unix()),
	}, nil
}

func (m *mockTokenVerifier) ExtractUserID(claims jwt.MapClaims) (uint, error) {
	return 1, nil
}

func TestMain(m *testing.M) {
	config.Load()
	os.Exit(m.Run())
}

func TestAuthMiddleware(t *testing.T) {
	verifier := &mockTokenVerifier{}
	// Setup a test Fiber app with your middleware and a dummy protected route
	app := fiber.New()
	app.Use(middleware.AuthMiddleware(verifier))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Generate a valid token for testing
	//validToken, _, err := auth.GenerateTokens(1)
	//if err != nil {
	//	t.Fatalf("failed to generate token: %v", err)
	//}
	validToken := "valid-token"

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
