package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amavis442/til-backend/internal/auth"
	"github.com/amavis442/til-backend/internal/handler"
	"github.com/amavis442/til-backend/internal/user"
	"github.com/gofiber/fiber/v2"
)

type mockUserService struct {
	ValidateCredentialsFunc func(username, password string) (bool, uint, error)
	GetByUsernameFunc       func(username string) (*user.User, error)
}

func (m *mockUserService) GetByUsername(username string) (*user.User, error) {
	return m.GetByUsernameFunc(username)
}

func (m *mockUserService) ValidateCredentials(username, password string) (bool, uint, error) {
	return m.ValidateCredentialsFunc(username, password)
}

func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name               string
		username           string
		password           string
		mockReturnValid    bool
		mockReturnUserID   uint
		mockReturnErr      error
		expectedStatus     int
		expectAccessToken  bool
		expectRefreshToken bool
	}{
		{
			name:               "valid credentials",
			username:           "admin",
			password:           "secret",
			mockReturnValid:    true,
			mockReturnUserID:   1,
			expectedStatus:     http.StatusOK,
			expectAccessToken:  true,
			expectRefreshToken: true,
		},
		{
			name:               "invalid credentials",
			username:           "admin",
			password:           "wrong",
			mockReturnValid:    false,
			mockReturnUserID:   0,
			expectedStatus:     http.StatusUnauthorized,
			expectAccessToken:  false,
			expectRefreshToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			mockSvc := &mockUserService{
				ValidateCredentialsFunc: func(username, password string) (bool, uint, error) {
					if username == tt.username && password == tt.password {
						return tt.mockReturnValid, tt.mockReturnUserID, tt.mockReturnErr
					}
					return false, 0, nil
				},
			}

			h := handler.NewAuthHandler(mockSvc)
			app.Post("/login", h.Login)

			body, _ := json.Marshal(map[string]string{
				"username": tt.username,
				"password": tt.password,
			})
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			var respBody map[string]string
			_ = json.NewDecoder(resp.Body).Decode(&respBody)

			if tt.expectAccessToken && respBody["access_token"] == "" {
				t.Error("Expected access_token in response")
			}
			if tt.expectRefreshToken && respBody["refresh_token"] == "" {
				t.Error("Expected refresh_token in response")
			}
			if !tt.expectAccessToken && respBody["access_token"] != "" {
				t.Error("Did not expect access_token in response")
			}
		})
	}
}

func TestRefreshTokenHandler(t *testing.T) {
	app := fiber.New()

	// Prepare mock service (not used in refresh handler, but required by constructor)
	mockSvc := &mockUserService{
		ValidateCredentialsFunc: func(username, password string) (bool, uint, error) {
			return false, 0, nil
		},
	}
	h := handler.NewAuthHandler(mockSvc)
	app.Post("/refresh", h.RefreshToken)

	// Generate valid refresh token
	userID := uint(42)
	_, refreshToken, err := auth.GenerateTokens(userID)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	tests := []struct {
		name           string
		refreshToken   string
		expectedStatus int
		expectAccess   bool
		expectRefresh  bool
	}{
		{
			name:           "valid refresh token",
			refreshToken:   refreshToken,
			expectedStatus: http.StatusOK,
			expectAccess:   true,
			expectRefresh:  true,
		},
		{
			name:           "invalid token format",
			refreshToken:   "not-a-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "empty token",
			refreshToken:   "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "non-refresh type token",
			refreshToken: func() string {
				access, _, _ := auth.GenerateTokens(userID)
				return access // returns an access token, not refresh
			}(),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{
				"refresh_token": tt.refreshToken,
			})
			req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Test request failed: %v", err)
			}
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedStatus == http.StatusOK {
				var data map[string]string
				_ = json.NewDecoder(resp.Body).Decode(&data)
				if tt.expectAccess && data["access_token"] == "" {
					t.Error("Expected access token in response")
				}
				if tt.expectRefresh && data["refresh_token"] == "" {
					t.Error("Expected refresh token in response")
				}
			}
		})
	}
}
