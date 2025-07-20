package user

import (
	"gorm.io/gorm"
)

type Repository interface {
	GetByUsername(username string) (*User, error)
	Create(user *User) error
	GetByID(id uint) (User, error)
	// add more DB methods here as needed
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByUsername(username string) (*User, error) {
	var user User
	result := r.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *repository) Create(user *User) error {
	return r.db.Create(&user).Error
}

func (r *repository) GetByID(id uint) (User, error) {
	var user User
	result := r.db.First(&user, id)
	return user, result.Error
}
