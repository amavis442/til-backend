package user

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username     string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"`
	Email        string `gorm:"not null"`
	Role         string `gorm:"type:varchar(50);not null;default:ROLE_USER"`
}
