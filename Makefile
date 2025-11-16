# Makefile for Go project

.PHONY: help build run test clean migrate-up migrate-down swagger docker-build docker-run

# Variables
APP_NAME := go-backend  # Change if your project name differs
VERSION := $(shell git describe --tags --always --dirty)
BUILD_DIR := ./bin
MAIN_PATH := ./cmd/api
MIGRATION_PATH := ./migrations

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod
GORUN := $(GOCMD) run

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $1, $2}'

install: ## Install dependencies
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Installing tools..."
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/air-verse/air@latest  # For hot reload

build: ## Build the application
	@echo "Building $(APP_NAME)..."
	$(GOBUILD) -ldflags "-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

run: ## Run the application
	$(GORUN) $(MAIN_PATH)/main.go

dev: ## Run in development mode with hot reload (requires air)
	air

# test: ## Run tests
# 	$(GOTEST) -v -race -coverprofile=coverage.out ./...
# 	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# test-integration: ## Run integration tests
# 	$(GOTEST) -v -tags=integration ./test/integration/...

# lint: ## Run linter
# 	golangci-lint run

swagger: ## Generate Swagger documentation
	swag init -g cmd/api/main.go -o internal/docs --parseInternal --parseDependency
	@echo "Swagger docs generated"

migrate-create: ## Create a new migration (use name=migration_name)
	migrate create -ext sql -dir $(MIGRATION_PATH) -seq $(name)

migrate-up: ## Run database migrations
	migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" up

migrate-down: ## Rollback database migrations
	migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" down

migrate-force: ## Force migration version (use version=N)
	migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" force $(version)

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	rm -rf coverage.out coverage.html
	@echo "Clean complete"

fmt: ## Format code
	go fmt ./...
	goimports -w .

vet: ## Run go vet
	go vet ./...

mod-update: ## Update dependencies
	$(GOMOD) get -u ./...
	$(GOMOD) tidy

security-scan: ## Run security scan
	gosec ./...

bench: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

coverage: ## Generate coverage report
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -func=coverage.out

all: clean install swagger build test ## Run all build steps

.DEFAULT_GOAL := help