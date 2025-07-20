package til

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

type Repository interface {
	GetAll(limit int, offset int) ([]TIL, error)
	Create(t TIL) error
	Update(til TIL) (TIL, error)
	GetByID(id uint) (TIL, error)
	Search(title, category string) ([]*TIL, error)
	FindOne(title, category string) (*TIL, error)
	Count() (int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetAll(limit int, offset int) ([]TIL, error) {
	var tils []TIL
	err := r.db.Order("created_at desc").Limit(limit).Offset(offset).Find(&tils).Error
	return tils, err
}

// Validation of t TIL is done in the service layer
func (r *repository) Create(t TIL) error {
	return r.db.Create(&t).Error
}

// Validation of t TIL is done in the service layer
func (r *repository) Update(til TIL) (TIL, error) {
	err := r.db.Omit("created_at, user_id").Save(&til).Error
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

// create a FindOne(title, category string)
func (r *repository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&TIL{}).Count(&count).Error
	return count, err
}

func (r *repository) FindOne(title, category string) (*TIL, error) {
	var til TIL

	if title == "" && category == "" {
		return nil, errors.New("both title and category are empty")
	}

	query := r.db
	if title != "" {
		query = query.Where("LOWER(title) LIKE ?", "%"+strings.ToLower(title)+"%")
	}
	if category != "" {
		query = query.Where("LOWER(category) = ?", strings.ToLower(category))
	}
	err := query.Limit(1).First(&til).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &til, nil
}
