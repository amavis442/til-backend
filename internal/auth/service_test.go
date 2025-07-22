package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/amavis442/til-backend/internal/auth"
)

type mockRepository struct {
	CreateFunc                     func(token *auth.RefreshToken) error
	FindRefreshTokenByUserIDFunc   func(userID uint) (*auth.RefreshToken, error)
	DeleteRefreshTokenFunc         func(token string) error
	DeleteRefreshTokenByUserIDFunc func(userID uint) error
}

func (m *mockRepository) Create(token *auth.RefreshToken) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(token)
	}
	return nil
}

func (m *mockRepository) FindRefreshTokenByUserID(userID uint) (*auth.RefreshToken, error) {
	if m.FindRefreshTokenByUserIDFunc != nil {
		return m.FindRefreshTokenByUserIDFunc(userID)
	}
	return nil, nil
}

func (m *mockRepository) DeleteRefreshToken(token string) error {
	if m.DeleteRefreshTokenFunc != nil {
		return m.DeleteRefreshTokenFunc(token)
	}
	return nil
}

func (m *mockRepository) DeleteRefreshTokenByUserID(userID uint) error {
	if m.DeleteRefreshTokenByUserIDFunc != nil {
		return m.DeleteRefreshTokenByUserIDFunc(userID)
	}
	return nil
}

func TestSaveRefreshToken(t *testing.T) {
	mockRepo := &mockRepository{
		CreateFunc: func(token *auth.RefreshToken) error {
			if token.UserID != 42 {
				t.Errorf("expected UserID to be 42, got %d", token.UserID)
			}
			if token.Token != "test.token.value" {
				t.Errorf("expected Token to be 'test.token.value', got %s", token.Token)
			}
			return nil
		},
	}

	// Inject a testable TokenExpiresAt function
	auth.TokenExpiresAt = func(token string) (*time.Time, error) {
		exp := time.Now().Add(1 * time.Hour)
		return &exp, nil
	}

	svc := auth.NewService(mockRepo)

	err := svc.SaveRefreshToken(42, "test.token.value")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSaveRefreshToken_TokenExpiresAtError(t *testing.T) {
	mockRepo := &mockRepository{}

	auth.TokenExpiresAt = func(token string) (*time.Time, error) {
		return nil, errors.New("parse error")
	}

	svc := auth.NewService(mockRepo)

	err := svc.SaveRefreshToken(1, "bad.token")
	if err == nil || err.Error() != "parse error" {
		t.Errorf("expected 'parse error', got %v", err)
	}
}

func TestFindRefreshTokenByUserID(t *testing.T) {
	mockRepo := &mockRepository{
		FindRefreshTokenByUserIDFunc: func(userID uint) (*auth.RefreshToken, error) {
			if userID != 123 {
				t.Errorf("expected userID 123, got %d", userID)
			}
			return &auth.RefreshToken{UserID: userID, Token: "some-token"}, nil
		},
	}

	svc := auth.NewService(mockRepo)

	token, err := svc.FindRefreshTokenByUserID(123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.UserID != 123 {
		t.Errorf("expected token for user 123, got %d", token.UserID)
	}
}

func TestDeleteRefreshToken(t *testing.T) {
	mockRepo := &mockRepository{
		DeleteRefreshTokenFunc: func(token string) error {
			if token != "delete-me" {
				t.Errorf("expected 'delete-me', got %s", token)
			}
			return nil
		},
	}

	svc := auth.NewService(mockRepo)

	err := svc.DeleteRefreshToken("delete-me")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
