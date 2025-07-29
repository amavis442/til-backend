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

func LoadEnv() {
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

	log.Printf("Loaded environment: %s", env)
}

func findProjectRoot() string {
	// Optional: search upward if needed
	return "."
}

func IsProduction() bool {
	env := os.Getenv("ENV")
	if env == "" {
		panic("no enviroment set in .env.local file ENV=dev or ENV=prod")
	}
	env = strings.ToLower(env)
	isProduction := true
	if env == "dev" {
		isProduction = false
	}
	slog.Info(fmt.Sprintf("Running in %v mode and flag is %v", env, isProduction))
	return isProduction
}
