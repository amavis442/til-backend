package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	root := findProjectRoot()

	// Always load .env (default config)
	_ = godotenv.Load(filepath.Join(root, ".env"))

	// Override with .env.local (if exists)
	_ = godotenv.Overload(filepath.Join(root, ".env.local"))

	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
		os.Setenv("ENV", env)
	}

	log.Printf("Loaded environment: %s", env)
}

func findProjectRoot() string {
	// Optional: search upward if needed
	return "."
}
