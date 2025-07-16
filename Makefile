# Makefile
include .env.local
export

# Default target
all: build run

# Build frontend and backend image
build:
	go build -o ./cmd/server/ ./cmd/server/main.go

# Shortcut for running the Go app locally (outside Docker)
run:
	go run ./cmd/server

# Migrate against production (external DB)
migrate-dev-up:
	migrate -path backend/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(PROD_DB_HOST):$(PROD_DB_PORT)/$(DB_NAME)_dev?sslmode=disable" up

migrate-dev-down:
	migrate -path backend/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(PROD_DB_HOST):$(PROD_DB_PORT)/$(DB_NAME)_dev?sslmode=disable" down

migrate-prod-up:
	migrate -path backend/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(PROD_DB_HOST):$(PROD_DB_PORT)/$(DB_NAME)?sslmode=disable" up

migrate-prod-down:
	migrate -path backend/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(PROD_DB_HOST):$(PROD_DB_PORT)/$(DB_NAME)?sslmode=disable" down

check:
	@echo "🔍 Checking required tools..."
	@command -v go >/dev/null 2>&1 || { echo "❌ Go is not installed."; exit 1; }
	@command -v migrate >/dev/null 2>&1 || { echo "❌ migrate CLI is not installed (https://github.com/golang-migrate/migrate)."; exit 1; }
	@echo "✅ All required tools are installed."

.PHONY: all build run migrate-dev-up migrate-prod-up migrate-dev-down migrate-prod-down check
