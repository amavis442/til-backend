package repository

import (
	"strings"

	"github.com/amavis442/til-backend/internal/domain"
	"gorm.io/gorm"
)

type tilRepository struct {
	db *gorm.DB
}

func NewTILRepository(db *gorm.DB) domain.TILRepository {
	return &tilRepository{db}
}

func (r *tilRepository) GetAll() ([]domain.TIL, error) {
	var tils []domain.TIL
	err := r.db.Order("created_at desc").Find(&tils).Error
	return tils, err
}

func (r *tilRepository) Create(t domain.TIL) error {
	return r.db.Create(&t).Error
}

func (r *tilRepository) Update(til domain.TIL) (domain.TIL, error) {
	err := r.db.Save(&til).Error
	return til, err
}

func (r *tilRepository) GetByID(id uint) (domain.TIL, error) {
	var til domain.TIL
	result := r.db.First(&til, id)
	return til, result.Error
}

func (r *tilRepository) Search(title, category string) ([]*domain.TIL, error) {
	var tils []*domain.TIL
	query := r.db

	if title != "" {
		query = query.Where("LOWER(title) LIKE ?", "%"+strings.ToLower(title)+"%")
	}
	if category != "" {
		query = query.Where("LOWER(category) = ?", strings.ToLower(category))
	}

	err := query.Order("created_at DESC").Find(&tils).Error
	return tils, err
}
