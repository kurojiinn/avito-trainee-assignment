package config

import "os"

// Config содержит всю конфигурацию приложения.
type Config struct {
	DB DBConfig
}

// DBConfig содержит параметры подключения к базе данных PostgreSQL.
type DBConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Name     string
}

// LoadConfig загружает конфигурацию из переменных окружения.
func LoadConfig() *Config {
	dbConfig := DBConfig{
		Host:     getEnv("POSTGRES_HOST", "localhost"),
		Port:     getEnv("POSTGRES_PORT", "5432"),
		Username: getEnv("POSTGRES_USER", "avito"),
		Password: getEnv("POSTGRES_PASSWORD", "avito"),
		Name:     getEnv("POSTGRES_DB", "mydatabase"),
	}

	return &Config{DB: dbConfig}
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
