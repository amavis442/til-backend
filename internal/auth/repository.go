package auth

import (
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	Create(refreshToken *RefreshToken) error
	FindRefreshTokenByUserID(userID uint) (*RefreshToken, error)
	DeleteRefreshToken(token string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(refreshToken *RefreshToken) error {
	return r.db.Create(&refreshToken).Error
}

func (r *repository) FindRefreshTokenByUserID(userID uint) (*RefreshToken, error) {
	var refreshToken RefreshToken
	result := r.db.Where("user_id = ?", userID).First(&refreshToken)
	return &refreshToken, result.Error
}

func (r *repository) DeleteRefreshToken(token string) error {
	if token == "" {
		return errors.New("refresh token must not be empty")
	}
	return r.db.Where("refresh_token = ?", token).Delete(&RefreshToken{}).Error
}
