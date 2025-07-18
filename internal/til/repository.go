package til

import (
	"strings"

	"gorm.io/gorm"
)

type Repository interface {
	GetAll() ([]TIL, error)
	Create(t TIL) error
	Update(til TIL) (TIL, error)
	GetByID(id uint) (TIL, error)
	Search(title, category string) ([]*TIL, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetAll() ([]TIL, error) {
	var tils []TIL
	err := r.db.Order("created_at desc").Find(&tils).Error
	return tils, err
}

func (r *repository) Create(t TIL) error {
	return r.db.Create(&t).Error
}

func (r *repository) Update(til TIL) (TIL, error) {
	err := r.db.Save(&til).Error
	return til, err
}

func (r *repository) GetByID(id uint) (TIL, error) {
	var til TIL
	result := r.db.First(&til, id)
	return til, result.Error
}

func (r *repository) Search(title, category string) ([]*TIL, error) {
	var tils []*TIL
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
