// Package db предоставляет функции для подключения к базе данных PostgreSQL.
// Использует стандартный драйвер database/sql для работы с PostgreSQL.
package db

import (
	"avito-assignment/internal/config"
	"database/sql"
	"fmt"
	"log"
)

// Connect устанавливает соединение с базой данных PostgreSQL.
// 
// Параметры:
//   - cfg: конфигурация подключения к БД (хост, порт, пользователь, пароль, имя БД)
//
// Возвращает:
//   - *sql.DB: объект соединения с БД
//
// Функция выполняет:
//   1. Формирует строку подключения (DSN) из конфигурации
//   2. Открывает соединение с БД
//   3. Проверяет доступность БД через Ping()
//   4. Завершает программу с ошибкой, если подключение не удалось
//
// Пример использования:
//   cfg := config.LoadConfig()
//   db := db.Connect(&cfg.DB)
//   defer db.Close()
func Connect(cfg *config.DBConfig) *sql.DB {
	// Формируем строку подключения (Data Source Name)
	// Формат: postgres://username:password@host:port/database?sslmode=disable
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Name,
	)

	// Открываем соединение с БД
	// sql.Open не устанавливает реальное соединение, только создает объект
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open DB connection: %v", err)
	}

	// Проверяем доступность БД через Ping()
	// Это реально устанавливает соединение и проверяет его работоспособность
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}

	log.Println("Connected to DB successfully")
	return db
}
