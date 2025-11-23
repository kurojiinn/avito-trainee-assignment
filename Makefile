export PATH := $(PATH):$(shell go env GOPATH)/bin

# ------------------------------
# Конфигурация БД
# ------------------------------
DB_USER := avito
DB_PASSWORD := avito
DB_HOST := localhost
DB_PORT := 5432
DB_NAME := mydatabase

# Строка подключения (DSN)
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

MIGRATION_DIR := "migrations"

# ------------------------------
# Основные команды
# ------------------------------

.PHONY: build
build:
	@echo "Building application..."
	@go build -o bin/api ./cmd/api

.PHONY: run
run: build
	@echo "Running application..."
	@./bin/api

.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out

.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	@docker-compose build

.PHONY: docker-up
docker-up:
	@echo "Starting services with docker-compose..."
	@docker-compose up -d

.PHONY: docker-down
docker-down:
	@echo "Stopping services..."
	@docker-compose down

.PHONY: docker-logs
docker-logs:
	@docker-compose logs -f api

.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/

# ------------------------------
# Миграции с goose
# ------------------------------

.PHONY: migrate-up
migrate-up:
	@echo "Applying migrations..."
	@goose -dir migrations postgres "$(DB_URL)" up

.PHONY: migrate-down
migrate-down:
	@echo "Rolling back last migration..."
	@goose -dir migrations postgres "$(DB_URL)" down

.PHONY: migrate-status
migrate-status:
	@echo "Migration status:"
	@goose -dir migrations postgres "$(DB_URL)" status

.PHONY: migrate-create
migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-create name=migration_name"; \
		exit 1; \
	fi
	@goose -dir $(MIGRATION_DIR) create $(name) sql

# ------------------------------
# Утилиты
# ------------------------------

.PHONY: install-deps
install-deps:
	@echo "Installing dependencies..."
	@go mod download
	@go install github.com/pressly/goose/v3/cmd/goose@latest

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...

.PHONY: vet
vet:
	@echo "Running go vet..."
	@go vet ./...

.PHONY: lint
lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

.PHONY: e2e-test
e2e-test:
	@echo "Running E2E tests..."
	@go test -v ./tests/...
