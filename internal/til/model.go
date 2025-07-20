package til

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

var (
	ErrValidation = errors.New("validation error") // ErrValidation is returned when a TIL entry fails validation.
	ErrDuplicate  = errors.New("duplicate entry")  // ErrDuplicate is returned when a TIL entry s a duplicate.
)

// TIL represents a "Today I Learned" entry.
type TIL struct {
	ID        uint           `json:"id" gorm:"primarykey"`    // Unique identifier for the TIL entry
	CreatedAt time.Time      `json:"created_at" gorm:"index"` // Timestamp when the entry was created
	UpdatedAt time.Time      // Timestamp when the entry was last updated
	DeletedAt gorm.DeletedAt `gorm:"index"`                                   // Soft delete timestamp (nullable)
	Title     string         `json:"title" gorm:"size:255;not null"`          // Title of the TIL entry
	Content   string         `json:"content" gorm:"type:text;not null"`       // Content or description of the TIL entry
	Category  string         `json:"category" gorm:"size:100;not null;index"` // Category of the TIL entry
	UserID    uint           `json:"user_id" gorm:"index;not null"`           // ID of the user who created the entry
}

func (t *TIL) Validate() error {
	if len(t.Title) == 0 || len(t.Title) > 255 {
		return fmt.Errorf("%w: title is required and must be at most 255 characters", ErrValidation)
	}
	if len(t.Content) == 0 {
		return fmt.Errorf("%w: content is required", ErrValidation)
	}
	if len(t.Category) == 0 || len(t.Category) > 100 {
		return fmt.Errorf("%w: category is required and must be at most 100 characters", ErrValidation)
	}
	if t.UserID == 0 {
		return fmt.Errorf("%w: user_id is required", ErrValidation)
	}
	return nil
}
