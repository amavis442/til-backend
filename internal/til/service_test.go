package til_test

import (
	"errors"
	"testing"

	"github.com/amavis442/til-backend/internal/til"
)

// --- Fake repository for testing ---
type fakeRepo struct {
	tList     []til.TIL
	createErr error
	updateRet til.TIL
	updateErr error
	getRet    til.TIL
	getErr    error
	searchRet []*til.TIL
	searchErr error
}

func (f *fakeRepo) GetAll() ([]til.TIL, error) {
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

func TestService_List(t *testing.T) {
	tests := []struct {
		name     string
		repoData []til.TIL
		wantLen  int
	}{
		{"empty list", []til.TIL{}, 0},
		{"one item", []til.TIL{{ID: 1, Title: "Go"}}, 1},
		{"multiple items", []til.TIL{{ID: 1}, {ID: 2}}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := til.NewService(&fakeRepo{tList: tt.repoData})
			got, err := svc.List()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.wantLen {
				t.Errorf("got %d items, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name      string
		input     til.TIL
		createErr error
		wantErr   bool
	}{
		{"success", til.TIL{Title: "Test"}, nil, false},
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
