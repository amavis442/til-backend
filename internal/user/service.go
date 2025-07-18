package user

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service interface {
	ValidateCredentials(username, password string) (bool, uint, error)
	GetByUsername(username string) (*User, error)
}

type service struct {
	repo Repository
}

func NewService(r Repository) Service {
	return &service{repo: r}
}

// ValidateCredentials compares the given password with the stored hash.
// Returns true and user ID if valid, false otherwise.
func (s *service) ValidateCredentials(username, password string) (bool, uint, error) {
	user, err := s.GetByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, 0, nil // user not found
		}
		return false, 0, err // DB error
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return false, 0, nil // invalid password
	}

	return true, user.ID, nil
}

func (s *service) GetByUsername(username string) (*User, error) {
	return s.repo.GetByUsername(username)
}
