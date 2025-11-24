package db

import (
	"avito-assignment/internal/config"
	"database/sql"
	"fmt"
	"log"
)

// Connect устанавливает соединение с базой данных PostgreSQL.
func Connect(cfg *config.DBConfig) *sql.DB {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Name,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open DB connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}

	log.Println("Connected to DB successfully")
	return db
}
