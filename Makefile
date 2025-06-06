# Makefile for assignment API

# Variables
APP_NAME := assignment-api
DOCKER_IMAGE := $(APP_NAME)
DOCKER_TAG := latest
GO_VERSION := 1.24

# Environment detection
ENV ?= development
ifeq ($(ENV),production)
	COMPOSE_FILE := docker-compose.prod.yaml
else ifeq ($(ENV),development)
	COMPOSE_FILE := docker-compose.dev.yaml
else
	COMPOSE_FILE := docker-compose.yaml
endif

# Colors for output
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

.PHONY: help
help: ## Show this help message
	@echo "$(GREEN)Available commands:$(RESET)"
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make $(GREEN)<target>$(RESET)\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(GREEN)%-15s$(RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(RESET)\n", substr($$0, 5) }' $(MAKEFILE_LIST)

##@ Development
.PHONY: dev
dev: ## Start development environment with hot reload
	@echo "$(GREEN)Starting development environment...$(RESET)"
	docker-compose -f docker-compose.dev.yaml up --build -d

.PHONY: dev-down
dev-down: ## Stop development environment
	@echo "$(YELLOW)Stopping development environment...$(RESET)"
	docker-compose -f docker-compose.dev.yaml down

.PHONY: dev-logs
dev-logs: ## Show development logs
	docker-compose -f docker-compose.dev.yaml logs -f

.PHONY: run
run: ## Run the application locally
	@echo "$(GREEN)Running application locally...$(RESET)"
	APP_ENV=development go run ./cmd/api

.PHONY: build
build: ## Build the application binary
	@echo "$(GREEN)Building application...$(RESET)"
	CGO_ENABLED=0 go build -o bin/$(APP_NAME) ./cmd/api

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(RESET)"
	rm -rf bin/
	go clean

##@ Testing
.PHONY: test
test: ## Run all tests
	@echo "$(GREEN)Running tests...$(RESET)"
	APP_ENV=test go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(RESET)"
	APP_ENV=test go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(RESET)"

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "$(GREEN)Running integration tests...$(RESET)"
	APP_ENV=test go test -v -tags=integration ./...

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "$(GREEN)Running benchmarks...$(RESET)"
	go test -bench=. -benchmem ./...

##@ Code Quality
.PHONY: lint
lint: ## Run linter
	@echo "$(GREEN)Running linter...$(RESET)"
	golangci-lint run

.PHONY: fmt
fmt: ## Format code
	@echo "$(GREEN)Formatting code...$(RESET)"
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "$(GREEN)Running go vet...$(RESET)"
	go vet ./...

.PHONY: mod-tidy
mod-tidy: ## Tidy go modules
	@echo "$(GREEN)Tidying go modules...$(RESET)"
	go mod tidy

.PHONY: check
check: fmt vet lint test ## Run all checks (format, vet, lint, test)

##@ Database
.PHONY: db-up
db-up: ## Start database only
	@echo "$(GREEN)Starting database...$(RESET)"
	docker-compose up -d postgres

.PHONY: db-down
db-down: ## Stop database
	@echo "$(YELLOW)Stopping database...$(RESET)"
	docker-compose down postgres

.PHONY: db-logs
db-logs: ## Show database logs
	docker-compose logs -f postgres

.PHONY: db-shell
db-shell: ## Connect to database shell
	docker-compose exec postgres psql -U postgres -d assignment-db

.PHONY: db-reset
db-reset: ## Reset database (WARNING: destroys data)
	@echo "$(RED)WARNING: This will destroy all data!$(RESET)"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ]
	docker-compose down postgres
	docker volume rm assignment_postgres_data 2>/dev/null || true
	docker-compose up -d postgres

.PHONY: gensql
gensql: ## Generate SQLBoiler models from database
	@echo "$(GREEN)Generating SQLBoiler models...$(RESET)"
	cd sqlboiler_config && sqlboiler psql -o ../internal/infrastructure/database/models

.PHONY: rebuild-api
rebuild-api: ## Force rebuild and recreate API container
	@echo "$(GREEN)Force rebuilding API container...$(RESET)"
	docker-compose up --build -d --force-recreate api

##@ Docker
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(GREEN)Building Docker image...$(RESET)"
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: docker-build-dev
docker-build-dev: ## Build Docker image for development
	@echo "$(GREEN)Building Docker image for development...$(RESET)"
	docker build --target development -t $(DOCKER_IMAGE):dev .

.PHONY: docker-build-prod
docker-build-prod: ## Build Docker image for production
	@echo "$(GREEN)Building Docker image for production...$(RESET)"
	docker build --target production -t $(DOCKER_IMAGE):prod .

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "$(GREEN)Running Docker container...$(RESET)"
	docker run -p 8080:8080 --env-file .env.development $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: up
up: ## Start application with docker-compose
	@echo "$(GREEN)Starting application ($(ENV) environment)...$(RESET)"
	docker-compose -f $(COMPOSE_FILE) up --build

.PHONY: down
down: ## Stop application
	@echo "$(YELLOW)Stopping application...$(RESET)"
	docker-compose -f $(COMPOSE_FILE) down

.PHONY: logs
logs: ## Show application logs
	docker-compose -f $(COMPOSE_FILE) logs -f

.PHONY: restart
restart: down up ## Restart application

##@ Production
.PHONY: prod
prod: ## Start production environment
	@echo "$(GREEN)Starting production environment...$(RESET)"
	ENV=production $(MAKE) up

.PHONY: prod-down
prod-down: ## Stop production environment
	@echo "$(YELLOW)Stopping production environment...$(RESET)"
	ENV=production $(MAKE) down

.PHONY: prod-logs
prod-logs: ## Show production logs
	ENV=production $(MAKE) logs

##@ Utilities
.PHONY: install-tools
install-tools: ## Install development tools
	@echo "$(GREEN)Installing development tools...$(RESET)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/air-verse/air@latest

.PHONY: deps
deps: ## Download dependencies
	@echo "$(GREEN)Downloading dependencies...$(RESET)"
	go mod download

.PHONY: update-deps
update-deps: ## Update dependencies
	@echo "$(GREEN)Updating dependencies...$(RESET)"
	go get -u ./...
	go mod tidy

.PHONY: env-example
env-example: ## Copy environment example file
	@echo "$(GREEN)Creating .env file from example...$(RESET)"
	cp .env.example .env
	@echo "$(YELLOW)Please edit .env file with your configuration$(RESET)"

.PHONY: health
health: ## Check application health
	@echo "$(GREEN)Checking application health...$(RESET)"
	@curl -f http://localhost:8080/health || echo "$(RED)Application is not healthy$(RESET)"

.PHONY: version
version: ## Show version information
	@echo "App: $(APP_NAME)"
	@echo "Go version: $(GO_VERSION)"
	@echo "Docker image: $(DOCKER_IMAGE):$(DOCKER_TAG)"

# Default target
.DEFAULT_GOAL := help