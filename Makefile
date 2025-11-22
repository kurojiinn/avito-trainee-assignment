export PATH := $(PATH):$(shell go env GOPATH)/bin

# ------------------------------
# Конфигурация БД
# ------------------------------
DB_USER := postgres
DB_PASSWORD := postgres
DB_HOST := localhost
DB_PORT := 5432
DB_NAME := mydatabase

# Строка подключения (DSN)
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable


MIGRATION_DIR := "migrations"

# ------------------------------
# Миграции с goose
# ------------------------------

# Применить все миграции
migrate-up:
	goose -dir migrations postgres "$(DB_URL)" up

# Откатить последнюю миграцию
migrate-down:
	goose -dir migrations postgres "$(DB_URL)" down

# Показать статус миграций
migrate-status:
	goose -dir migrations postgres "$(DB_URL)" status

# Создать новую миграцию
migrate-create:
	goose -dir $(MIGRATION_DIR) create $(name) sql
