package repository_tests

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// SetupTestDB создает тестовую базу данных и применяет миграции.
// Использует уникальное имя БД для каждого теста, чтобы избежать конфликтов
// при параллельном запуске тестов.
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Получаем параметры подключения из переменных окружения или используем значения по умолчанию
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	user := getEnv("POSTGRES_USER", "avito")
	password := getEnv("POSTGRES_PASSWORD", "avito")

	// Генерируем уникальное имя БД для каждого теста
	// Используем UUID и timestamp для гарантии уникальности
	testID := uuid.New().String()[:8]
	timestamp := time.Now().UnixNano()
	dbName := fmt.Sprintf("test_db_%s_%d", testID, timestamp)

	// Подключаемся к postgres для создания тестовой БД
	adminDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable", user, password, host, port)
	adminDB, err := sql.Open("postgres", adminDSN)
	if err != nil {
		t.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer adminDB.Close()

	// Принудительно закрываем все соединения с БД, если она существует
	// Это необходимо, чтобы можно было удалить БД
	_, err = adminDB.Exec(fmt.Sprintf(`
		SELECT pg_terminate_backend(pg_stat_activity.pid)
		FROM pg_stat_activity
		WHERE pg_stat_activity.datname = '%s'
		AND pid <> pg_backend_pid()
	`, dbName))
	// Игнорируем ошибку, если БД не существует

	// Удаляем БД, если она существует
	_, err = adminDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// Создаем новую тестовую БД
	_, err = adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Подключаемся к тестовой БД
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbName)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Настраиваем параметры соединения для тестов
	db.SetConnMaxLifetime(time.Minute * 5)
	db.SetMaxOpenConns(1) // Ограничиваем количество соединений для тестов
	db.SetMaxIdleConns(1)

	// Применяем миграции
	if err := applyMigrations(db); err != nil {
		db.Close()
		// Пытаемся удалить БД при ошибке миграции
		adminDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
		t.Fatalf("Failed to apply migrations: %v", err)
	}

	// Регистрируем автоматическую очистку БД после завершения теста
	t.Cleanup(func() {
		// Закрываем все соединения с БД
		db.Close()
		// Даем время на закрытие соединений
		time.Sleep(100 * time.Millisecond)

		// Подключаемся к postgres для удаления тестовой БД
		cleanupAdminDB, err := sql.Open("postgres", adminDSN)
		if err != nil {
			return
		}
		defer cleanupAdminDB.Close()

		// Принудительно закрываем все соединения с тестовой БД
		cleanupAdminDB.Exec(fmt.Sprintf(`
			SELECT pg_terminate_backend(pg_stat_activity.pid)
			FROM pg_stat_activity
			WHERE pg_stat_activity.datname = '%s'
			AND pid <> pg_backend_pid()
		`, dbName))

		// Даем время на закрытие соединений
		time.Sleep(100 * time.Millisecond)

		// Удаляем тестовую БД
		cleanupAdminDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	})

	return db
}

// CleanupTestDB закрывает соединение с тестовой базой данных.
// Удаление БД происходит автоматически через t.Cleanup в SetupTestDB.
// Эта функция оставлена для обратной совместимости.
func CleanupTestDB(t *testing.T, db *sql.DB) {
	t.Helper()
	if db != nil {
		db.Close()
	}
}

// applyMigrations применяет миграции к БД
func applyMigrations(db *sql.DB) error {
	// Создаем расширение для UUID
	if _, err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`); err != nil {
		return err
	}

	// Создаем тип для статуса PR
	if _, err := db.Exec(`DO $$ BEGIN CREATE TYPE pr_status AS ENUM ('OPEN','MERGED'); EXCEPTION WHEN duplicate_object THEN null; END $$;`); err != nil {
		return err
	}

	// Создаем таблицы
	queries := []string{
		`CREATE TABLE IF NOT EXISTS teams (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name TEXT NOT NULL UNIQUE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			username TEXT NOT NULL UNIQUE,
			team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS pull_requests (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			pull_request_name TEXT NOT NULL,
			author_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
			status pr_status NOT NULL DEFAULT 'OPEN',
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
			merged_at TIMESTAMP WITH TIME ZONE NULL
		)`,
		`CREATE TABLE IF NOT EXISTS pr_reviewers (
			pr_id UUID NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
			reviewer_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
			assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
			PRIMARY KEY (pr_id, reviewer_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_team_id ON users(team_id)`,
		`CREATE INDEX IF NOT EXISTS idx_pr_author ON pull_requests(author_id)`,
		`CREATE INDEX IF NOT EXISTS idx_pr_reviewers_reviewer ON pr_reviewers(reviewer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_pr_reviewers_pr ON pr_reviewers(pr_id)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
