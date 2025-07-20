package til_test

import (
	"errors"
	"testing"

	"github.com/amavis442/til-backend/internal/til"
)

// --- Fake repository for testing ---
type fakeRepo struct {
	tList     []til.TIL
	tListErr  error
	createErr error
	updateRet til.TIL
	updateErr error
	getRet    til.TIL
	getErr    error
	searchRet []*til.TIL
	searchErr error
	findRet   *til.TIL
	findErr   error
}

func (f *fakeRepo) GetAll() ([]til.TIL, error) {
	if f.tListErr != nil {
		return nil, f.tListErr
	}
	return f.tList, nil
}
func (f *fakeRepo) Create(t til.TIL) error {
	return f.createErr
}
func (f *fakeRepo) Update(t til.TIL) (til.TIL, error) {
	return f.updateRet, f.updateErr
}
func (f *fakeRepo) GetByID(id uint) (til.TIL, error) {
	return f.getRet, f.getErr
}
func (f *fakeRepo) Search(title, category string) ([]*til.TIL, error) {
	return f.searchRet, f.searchErr
}

func (f *fakeRepo) FindOne(title, category string) (*til.TIL, error) {
	return f.findRet, f.findErr
}

// --- Additional fakeRepo for spying ---
type spyRepo struct {
	fakeRepo
	findOneCalled   bool
	findOneTitle    string
	findOneCategory string
	createCalled    bool
}

func (s *spyRepo) FindOne(title, category string) (*til.TIL, error) {
	s.findOneCalled = true
	s.findOneTitle = title
	s.findOneCategory = category
	return s.findRet, s.findErr
}

func (s *spyRepo) Create(t til.TIL) error {
	s.createCalled = true
	return s.createErr
}

func TestService_List(t *testing.T) {
	// Helper function to compare slices of til.TIL for equality.
	equalTILs := func(a, b []til.TIL) bool {
		if len(a) != len(b) {
			return false
		}
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}

	tests := []struct {
		name     string
		repoData []til.TIL
		wantLen  int
	}{
		// Test with an empty list; expect 0 items returned.
		{"empty list", []til.TIL{}, 0},
		// Test with a non-nil, empty slice; expect 0 items returned.
		{"non-nil empty slice", make([]til.TIL, 0), 0},
		// Test with a single item in the repo; expect 1 item returned.
		{"one item", []til.TIL{{ID: 1, Title: "Go", UserID: 1}}, 1},
		// Test with multiple items in the repo; expect correct count.
		{"multiple items", []til.TIL{{ID: 1, UserID: 1}, {ID: 2, UserID: 1}}, 2},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			svc := til.NewService(&fakeRepo{tList: tt.repoData})
			got, err := svc.List()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.wantLen {
				t.Errorf("got %d items, want %d", len(got), tt.wantLen)
			}
			if !equalTILs(got, tt.repoData) {
				t.Errorf("got %+v, want %+v", got, tt.repoData)
			}
		})
	}
	t.Run("repo error", func(t *testing.T) {
		repoErr := errors.New("repo failure")
		svc := til.NewService(&fakeRepo{tListErr: repoErr})
		_, err := svc.List()
		if !errors.Is(err, repoErr) {
			t.Errorf("expected error %v, got %v", repoErr, err)
		}
	})
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name      string
		input     til.TIL
		createErr error
		wantErr   bool
	}{
		{"success", til.TIL{Title: "Test", Content: "Test Content", UserID: 1}, nil, false},
		{"create error", til.TIL{Title: "Fail"}, errors.New("fail"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := til.NewService(&fakeRepo{createErr: tt.createErr})
			err := svc.Create(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error: %v, wantErr: %v", err, tt.wantErr)
			}
		})
	}
}

// Service returns ErrDuplicate when a TIL entry with the same title and category already exists
func TestService_Create_ReturnsErrDuplicateIfExists(t *testing.T) {
	duplicate := &til.TIL{ID: 1, Title: "Go", Category: "Programming"}
	repo := &fakeRepo{findRet: duplicate}
	svc := til.NewService(repo)
	err := svc.Create(til.TIL{Title: "Go", Category: "Programming"})
	if !errors.Is(err, til.ErrDuplicate) {
		t.Errorf("expected ErrDuplicate, got %v", err)
	}
}

// Service calls repository FindOne with correct title and category parameters
func TestService_Create_CallsFindOneWithCorrectParams(t *testing.T) {
	repo := &spyRepo{}
	svc := til.NewService(repo)
	title := "Go"
	category := "Programming"
	_ = svc.Create(til.TIL{Title: title, Category: category})
	if !repo.findOneCalled {
		t.Error("expected FindOne to be called")
	}
	if repo.findOneTitle != title || repo.findOneCategory != category {
		t.Errorf("FindOne called with title=%q, category=%q; want title=%q, category=%q", repo.findOneTitle, repo.findOneCategory, title, category)
	}
}

// Service propagates repository FindOne errors without modification
func TestService_Create_PropagatesFindOneError(t *testing.T) {
	repoErr := errors.New("find error")
	repo := &fakeRepo{findErr: repoErr}
	svc := til.NewService(repo)
	err := svc.Create(til.TIL{Title: "Go", Category: "Programming"})
	if !errors.Is(err, repoErr) {
		t.Errorf("expected error %v, got %v", repoErr, err)
	}
}

// Service does not call repository Create if a duplicate TIL entry is found
func TestService_Create_DoesNotCallCreateOnDuplicate(t *testing.T) {
	duplicate := &til.TIL{ID: 1, Title: "Go", Category: "Programming"}
	repo := &spyRepo{fakeRepo: fakeRepo{
		findRet: duplicate,
	}}
	svc := til.NewService(repo)
	_ = svc.Create(til.TIL{Title: "Go", Category: "Programming"})
	if repo.createCalled {
		t.Error("expected Create not to be called when duplicate exists")
	}
}

// Service handles nil repository implementation gracefully
func TestService_Create_NilRepository(t *testing.T) {
	var nilRepo *fakeRepo = nil
	svc := til.NewService(nilRepo)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic or error when repository is nil, but got none")
		}
	}()
	_ = svc.Create(til.TIL{Title: "Go", Category: "Programming"})
}

// Service returns error if repository Create returns an unexpected error
func TestService_Create_HandlesUnexpectedCreateError(t *testing.T) {
	createErr := errors.New("unexpected create error")
	repo := &fakeRepo{createErr: createErr}
	svc := til.NewService(repo)
	err := svc.Create(til.TIL{Title: "Go", Category: "Programming"})
	if !errors.Is(err, createErr) {
		t.Errorf("expected error %v, got %v", createErr, err)
	}
}

func TestService_GetByID(t *testing.T) {
	expected := til.TIL{ID: 1, Title: "Go"}
	svc := til.NewService(&fakeRepo{getRet: expected})

	t.Run("success", func(t *testing.T) {
		got, err := svc.GetByID(1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != expected.ID {
			t.Errorf("got ID %d, want %d", got.ID, expected.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc := til.NewService(&fakeRepo{getErr: errors.New("not found")})
		_, err := svc.GetByID(999)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestService_Update(t *testing.T) {
	tests := []struct {
		name      string
		input     til.TIL
		updateRet til.TIL
		updateErr error
		wantID    uint
		wantErr   bool
	}{
		{
			name:      "successful update",
			input:     til.TIL{ID: 1, Title: "Updated"},
			updateRet: til.TIL{ID: 1, Title: "Updated"},
			wantID:    1,
			wantErr:   false,
		},
		{
			name:      "update failure",
			input:     til.TIL{ID: 2, Title: "Fail"},
			updateErr: errors.New("update failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := til.NewService(&fakeRepo{
				updateRet: tt.updateRet,
				updateErr: tt.updateErr,
			})
			got, err := svc.Update(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("expected error=%v, got err=%v", tt.wantErr, err)
			}
			if !tt.wantErr && got.ID != tt.wantID {
				t.Errorf("expected ID=%d, got ID=%d", tt.wantID, got.ID)
			}
		})
	}
}

func TestService_Search(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		category    string
		searchRet   []*til.TIL
		searchErr   error
		wantCount   int
		expectError bool
	}{
		{
			name:      "no results",
			title:     "none",
			category:  "none",
			searchRet: []*til.TIL{},
			wantCount: 0,
		},
		{
			name:      "single result",
			title:     "Go",
			category:  "programming",
			searchRet: []*til.TIL{{ID: 1, Title: "Go"}},
			wantCount: 1,
		},
		{
			name:        "search error",
			title:       "error",
			searchErr:   errors.New("search failed"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := til.NewService(&fakeRepo{
				searchRet: tt.searchRet,
				searchErr: tt.searchErr,
			})
			got, err := svc.Search(tt.title, tt.category)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error=%v, got err=%v", tt.expectError, err)
			}
			if !tt.expectError && len(got) != tt.wantCount {
				t.Errorf("expected count=%d, got=%d", tt.wantCount, len(got))
			}
		})
	}
}
