package til

type Service interface {
	List() ([]TIL, error)
	Create(t TIL) error
	Update(t TIL) (TIL, error)
	GetByID(id uint) (TIL, error)
	Search(title, category string) ([]*TIL, error)
}

type service struct {
	repo Repository
}

func NewService(r Repository) Service {
	return &service{r}
}

func (uc *service) List() ([]TIL, error) {
	return uc.repo.GetAll()
}

func (uc *service) Create(t TIL) error {
	return uc.repo.Create(t)
}

func (u *service) GetByID(id uint) (TIL, error) {
	return u.repo.GetByID(id)
}

func (uc *service) Update(til TIL) (TIL, error) {
	return uc.repo.Update(til)
}

func (u *service) Search(title, category string) ([]*TIL, error) {
	return u.repo.Search(title, category)
}
