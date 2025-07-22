package auth

import (
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	Create(refreshToken *RefreshToken) error
	FindRefreshTokenByUserID(userID uint) (*RefreshToken, error)
	DeleteRefreshToken(token string) error
	DeleteRefreshTokenByUserID(userID uint) error
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
	result := r.db.Where("user_id = ?", userID).Last(&refreshToken)
	return &refreshToken, result.Error
}

func (r *repository) DeleteRefreshToken(token string) error {
	if token == "" {
		return errors.New("refresh token must not be empty")
	}
	return r.db.Where("refresh_token = ?", token).Delete(&RefreshToken{}).Error
}

func (r *repository) DeleteRefreshTokenByUserID(userID uint) error {
	if userID == 0 {
		return errors.New("user id must not be empty cannot remove refresh token")
	}
	return r.db.Where("user_id = ?", userID).Delete(&RefreshToken{}).Error
}
