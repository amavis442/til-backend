package usecase

import "github.com/amavis442/til-backend/internal/domain"

type tilUsecase struct {
	repo domain.TILRepository
}

func NewTILUsecase(r domain.TILRepository) domain.TILUsecase {
	return &tilUsecase{r}
}

func (uc *tilUsecase) List() ([]domain.TIL, error) {
	return uc.repo.GetAll()
}

func (uc *tilUsecase) Create(t domain.TIL) error {
	return uc.repo.Create(t)
}

func (u *tilUsecase) GetByID(id uint) (domain.TIL, error) {
	return u.repo.GetByID(id)
}

func (uc *tilUsecase) Update(til domain.TIL) (domain.TIL, error) {
	return uc.repo.Update(til)
}

func (u *tilUsecase) Search(title, category string) ([]*domain.TIL, error) {
	return u.repo.Search(title, category)
}
