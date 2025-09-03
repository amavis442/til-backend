package config

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	CORSAllowedOrigin string
	ENV               string
	DB_DSN            string
	PORT              string
	// Add more vars here: DB_URL, PORT, etc.
}

var requiredEnvVars = []string{
	"CORS_ALLOWED_ORIGIN",
	"ENV",
	"DB_DSN", "PORT",
}

func validateRequiredEnvVars(keys []string) {
	for _, key := range keys {
		if os.Getenv(key) == "" {
			log.Fatalf("Missing required env var: %s", key)
		}
	}
}

func Load() Config {
	root := findProjectRoot()

	// Always load .env (default config)
	_ = godotenv.Load(filepath.Join(root, ".env"))

	// Override with .env.local (if exists)
	envPath := filepath.Join(root, ".env.local")
	if err := godotenv.Overload(envPath); err != nil {
		log.Printf("No .env.local found at %s — relying on system env", envPath)
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
		os.Setenv("ENV", env)
	}

	validateRequiredEnvVars(requiredEnvVars)

	slog.Info(fmt.Sprintf("Loaded environment: %s", env))

	return Config{
		CORSAllowedOrigin: os.Getenv("CORS_ALLOWED_ORIGIN"),
		ENV:               env,
		DB_DSN:            os.Getenv("DB_DSN"),
		PORT:              os.Getenv("PORT"),
	}
}

func findProjectRoot() string {
	//return "."
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			break
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			panic(fmt.Errorf("go.mod not found"))
		}
		currentDir = parent
	}

	return currentDir
}

func IsProduction() bool {
	env := os.Getenv("ENV")
	if env == "" {
		log.Fatal("no environment set in .env.local file ENV=dev or ENV=prod")
	}
	env = strings.ToLower(env)
	isProduction := true
	if env == "dev" {
		isProduction = false
	}

	return isProduction
}
