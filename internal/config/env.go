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
	err := godotenv.Overload(filepath.Join(root, ".env.local"))
	if err != nil {
		log.Fatal("Could not find .env or .env.local in " + envPath)
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
	// Optional: search upward if needed
	return "."
}

func IsProduction() bool {
	env := os.Getenv("ENV")
	if env == "" {
		log.Fatal("no enviroment set in .env.local file ENV=dev or ENV=prod")
	}
	env = strings.ToLower(env)
	isProduction := true
	if env == "dev" {
		isProduction = false
	}

	return isProduction
}
