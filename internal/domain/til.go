package domain

import "time"

type TIL struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
}

type TILRepository interface {
	GetAll() ([]TIL, error)
	Create(t TIL) error
	Update(t TIL) (TIL, error)
	GetByID(id uint) (TIL, error)
	Search(title, category string) ([]*TIL, error)
}

type TILUsecase interface {
	List() ([]TIL, error)
	Create(t TIL) error
	Update(t TIL) (TIL, error)
	GetByID(id uint) (TIL, error)
	Search(title, category string) ([]*TIL, error)
}
