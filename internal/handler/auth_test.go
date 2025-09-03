package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/amavis442/til-backend/internal/auth"
	"github.com/amavis442/til-backend/internal/config"
	"github.com/amavis442/til-backend/internal/handler"
	"github.com/amavis442/til-backend/internal/middleware"
	"github.com/amavis442/til-backend/internal/user"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type mockUserService struct {
	ValidateCredentialsFunc func(username, password string) (bool, uint, error)
	GetByUsernameFunc       func(username string) (*user.User, error)
	RegisterFunc            func(username, email, password string) error
	UserExistsFunc          func(userID uint) (bool, error)
	UpdatePasswordFunc      func(userID uint, password string) error
}

func (m *mockUserService) GetByUsername(username string) (*user.User, error) {
	return m.GetByUsernameFunc(username)
}

func (m *mockUserService) ValidateCredentials(username, password string) (bool, uint, error) {
	return m.ValidateCredentialsFunc(username, password)
}

func (m *mockUserService) Register(username, email, password string) error {
	return m.RegisterFunc(username, email, password)
}

func (m *mockUserService) UserExists(userID uint) (bool, error) {
	return m.UserExistsFunc(userID)
}

func (m *mockUserService) UpdatePassword(userID uint, password string) error {
	return m.UpdatePasswordFunc(userID, password)
}

type mockUserRepository struct {
	GetByIDFunc       func(id uint) (user.User, error)
	UpdateFunc        func(user *user.User) error
	CreateFunc        func(user *user.User) error
	GetByUsernameFunc func(username string) (*user.User, error)
}

func (m *mockUserRepository) GetByID(id uint) (user.User, error) {
	return m.GetByIDFunc(id)
}

func (m *mockUserRepository) Update(user *user.User) error {
	return m.UpdateFunc(user)
}

func (m *mockUserRepository) Create(user *user.User) error {
	return m.CreateFunc(user)
}

func (m *mockUserRepository) GetByUsername(username string) (*user.User, error) {
	return m.GetByUsernameFunc(username)
}

type mockRefreshTokenService struct {
	CreateFunc                     func(userID uint, token string) error
	FindRefreshTokenByUserIDFunc   func(userID uint) (*auth.RefreshToken, error)
	DeleteRefreshTokenFunc         func(token string) error
	DeleteRefreshTokenByUserIDFunc func(userID uint) error
}

func (m *mockRefreshTokenService) SaveRefreshToken(userID uint, token string) error {
	return m.CreateFunc(userID, token)
}

func (m *mockRefreshTokenService) FindRefreshTokenByUserID(userID uint) (*auth.RefreshToken, error) {
	return m.FindRefreshTokenByUserIDFunc(userID)
}

func (m *mockRefreshTokenService) DeleteRefreshToken(token string) error {
	return m.DeleteRefreshTokenFunc(token)
}

func (m *mockRefreshTokenService) DeleteRefreshTokenByUserID(userID uint) error {
	return m.DeleteRefreshTokenByUserIDFunc(userID)
}

func TestMain(m *testing.M) {
	root := "../../"
	config.Load()

	if err := auth.InitJWTKeys(root); err != nil {
		log.Fatalf("failed to load keys: %v", err)
	}

	// Run the tests
	os.Exit(m.Run())
}

func TestEnvLoaded(t *testing.T) {
	t.Log("JWT_PRIVATE_KEY_PATH =", os.Getenv("JWT_PRIVATE_KEY_PATH"))
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

			mockRefreshTokenSvc := &mockRefreshTokenService{
				CreateFunc: func(userID uint, token string) error {
					return nil
				},
				DeleteRefreshTokenByUserIDFunc: func(userID uint) error {
					return nil
				},
			}

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			h := handler.NewAuthHandler(mockSvc, mockRefreshTokenSvc, logger)
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
	userID := uint(42)

	mockRefreshTokenSvc := &mockRefreshTokenService{
		FindRefreshTokenByUserIDFunc: func(userID uint) (*auth.RefreshToken, error) {
			token, err := auth.GenerateRefreshToken(userID)
			if err != nil {
				t.Fatalf("Failed to generate refresh token: %v", err)
			}

			refreshToken := auth.RefreshToken{
				UserID: userID,
				Token:  token,
			}

			return &refreshToken, nil
		},
		DeleteRefreshTokenFunc: func(token string) error {
			return nil
		},
		CreateFunc: func(userID uint, token string) error {
			return nil
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := handler.NewAuthHandler(mockSvc, mockRefreshTokenSvc, logger)
	app.Post("/refresh", h.RefreshToken)

	// Generate valid refresh token
	//userID := uint(42)
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
					t.Error("Expected access token in response", data)
				}
				if tt.expectRefresh && data["refresh_token"] == "" {
					t.Error("Expected refresh token in response")
				}
			}
		})
	}
}

func TestRegisterHandler(t *testing.T) {
	type testCase struct {
		name               string
		body               string
		setupMock          func(svc *mockUserService)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name: "valid registration",
			body: `{"username":"newuser", "email":"new@example.com", "password":"password123"}`,
			setupMock: func(svc *mockUserService) {
				svc.RegisterFunc = func(username, email, password string) error {
					assert.Equal(t, "newuser", username)
					assert.Equal(t, "new@example.com", email)
					assert.Equal(t, "password123", password)
					return nil
				}
			},
			expectedStatusCode: fiber.StatusCreated,
		},
		{
			name: "missing fields",
			body: `{"username":"", "email":"", "password":""}`,
			setupMock: func(svc *mockUserService) {
				svc.RegisterFunc = func(username, email, password string) error {
					t.Error("Register should not be called on invalid input")
					return nil
				}
			},
			expectedStatusCode: fiber.StatusBadRequest,
		},
		{
			name: "internal error",
			body: `{"username":"failuser", "email":"fail@example.com", "password":"password123"}`,
			setupMock: func(svc *mockUserService) {
				svc.RegisterFunc = func(username, email, password string) error {
					return errors.New("db error")
				}
			},
			expectedStatusCode: fiber.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			mockSvc := &mockUserService{}
			tc.setupMock(mockSvc)

			mockRefeshTokenSvc := &mockRefreshTokenService{}

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))

			handler := handler.NewAuthHandler(mockSvc, mockRefeshTokenSvc, logger)

			app.Post("/register", handler.Register)

			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatusCode, resp.StatusCode)
		})
	}
}

type mockTokenVerifier struct{}

func (m *mockTokenVerifier) Verify(tokenStr string) (jwt.MapClaims, error) {
	return jwt.MapClaims{
		"sub": "1",
		"typ": "access",
		"exp": float64(time.Now().Add(15 * time.Minute).Unix()),
	}, nil
}

func (m *mockTokenVerifier) ExtractUserID(claims jwt.MapClaims) (uint, error) {
	return 1, nil
}

func TestUpdatePasswordHandler_WithAuth(t *testing.T) {
	app := fiber.New()

	mockRepo := &mockUserRepository{
		GetByIDFunc: func(id uint) (user.User, error) {
			return user.User{Model: gorm.Model{ID: 1}, Username: "testuser", PasswordHash: "oldhash"}, nil
		},
		UpdateFunc: func(u *user.User) error {
			assert.Equal(t, uint(1), u.ID)
			assert.NotEqual(t, "oldhash", u.PasswordHash)
			return nil
		},
	}

	svc := user.NewService(mockRepo)
	mockRefreshTokenSvc := &mockRefreshTokenService{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := handler.NewAuthHandler(svc, mockRefreshTokenSvc, logger)

	verifier := &mockTokenVerifier{}
	app.Post("/api/change-password", middleware.AuthMiddleware(verifier), handler.UpdatePassword)

	body := `{"password":"newpassword123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/change-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer dummy-token")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}
