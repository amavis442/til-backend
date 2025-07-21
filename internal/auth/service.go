package auth

import (
	"errors"
)

type Service interface {
	SaveRefreshToken(userID uint, token string) error
	FindRefreshTokenByUserID(userID uint) (*RefreshToken, error)
	DeleteRefreshToken(token string) error
}

type service struct {
	repo Repository
}

func NewService(r Repository) Service {
	return &service{repo: r}
}

// Create implements Service.
func (s *service) SaveRefreshToken(userID uint, token string) error {
	refreshTokenExpiresAt, err := TokenExpiresAt(token)
	if err != nil {
		return err
	}

	var refreshToken RefreshToken
	refreshToken.UserID = userID
	refreshToken.Token = token
	refreshToken.ExpiresAt = *refreshTokenExpiresAt
	return s.repo.Create(&refreshToken)
}

// GetTokenByUserID implements Service.
func (s *service) FindRefreshTokenByUserID(userID uint) (*RefreshToken, error) {
	if userID == 0 {
		return nil, errors.New("invalid userID")
	}
	return s.repo.FindRefreshTokenByUserID(userID)
}

func (s *service) DeleteRefreshToken(token string) error {
	return s.repo.DeleteRefreshToken(token)
}
