package til

type Service interface {
	List(limit int, offset int) ([]TIL, error)
	ListWithCount(limit int, offset int) ([]TIL, int64, error)
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

func (uc *service) List(limit int, offset int) ([]TIL, error) {
	return uc.repo.GetAll(limit, offset)
}

func (uc *service) ListWithCount(limit int, offset int) ([]TIL, int64, error) {
	tils, err := uc.repo.GetAll(limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := uc.repo.Count()
	if err != nil {
		return nil, 0, err
	}
	return tils, total, nil
}

func (uc *service) Create(t TIL) error {
	til, err := uc.repo.FindOne(t.Title, t.Category)
	if err != nil {
		return err
	}
	if til != nil {
		return ErrDuplicate
	}
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
