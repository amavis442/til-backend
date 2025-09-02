package user

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service interface {
	ValidateCredentials(username, password string) (bool, uint, error)
	GetByUsername(username string) (*User, error)
	Register(username, email, password string) error
	UserExists(userID uint) (bool, error)
	UpdatePassword(userID uint, password string) error
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

// ExistsByID checks if a user with the given userID exists and returns true if found, otherwise false.
func (s *service) UserExists(userID uint) (bool, error) {
	_, err := s.repo.GetByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *service) Register(username, email, password string) error {
	// Check if user already exists
	existing, err := s.repo.GetByUsername(username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existing != nil {
		return fmt.Errorf("username already taken")
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Create user
	user := &User{
		Username:     username,
		PasswordHash: string(hashed),
		Email:        email,
		Role:         "ROLE_USER", // Default role
	}

	err = s.repo.Create(user)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) UpdatePassword(userID uint, password string) error {
	user, err := s.repo.GetByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("User not found")
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashed)

	s.repo.Update(&user)

	return nil
}
