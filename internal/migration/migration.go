package migration

import (
	"log"

	"github.com/amavis442/til-backend/internal/domain"
	"gorm.io/gorm"
)

func Run(db *gorm.DB) {
	err := db.AutoMigrate(&domain.TIL{})
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}
