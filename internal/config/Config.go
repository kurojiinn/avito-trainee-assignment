// Package config предоставляет функции для загрузки конфигурации приложения.
// Конфигурация загружается из переменных окружения с значениями по умолчанию.
package config

import "os"

// Config содержит всю конфигурацию приложения.
// Включает настройки подключения к базе данных и другие параметры.
type Config struct {
	// DB - конфигурация подключения к базе данных PostgreSQL
	DB DBConfig
}

// DBConfig содержит параметры подключения к базе данных PostgreSQL.
// Все значения могут быть переопределены через переменные окружения.
type DBConfig struct {
	// Host - адрес хоста PostgreSQL (по умолчанию: localhost)
	Host string
	
	// Port - порт PostgreSQL (по умолчанию: 5432)
	Port string
	
	// Username - имя пользователя для подключения к БД (по умолчанию: avito)
	Username string
	
	// Password - пароль для подключения к БД (по умолчанию: avito)
	Password string
	
	// Name - имя базы данных (по умолчанию: mydatabase)
	Name string
}

// LoadConfig загружает конфигурацию из переменных окружения.
// 
// Использует следующие переменные окружения:
//   - POSTGRES_HOST - хост PostgreSQL (по умолчанию: localhost)
//   - POSTGRES_PORT - порт PostgreSQL (по умолчанию: 5432)
//   - POSTGRES_USER - пользователь БД (по умолчанию: avito)
//   - POSTGRES_PASSWORD - пароль БД (по умолчанию: avito)
//   - POSTGRES_DB - имя БД (по умолчанию: mydatabase)
//
// Возвращает:
//   - *Config: объект конфигурации с загруженными значениями
//
// Пример использования:
//   cfg := config.LoadConfig()
//   db := db.Connect(&cfg.DB)
func LoadConfig() *Config {
	// Загружаем конфигурацию БД из переменных окружения
	// Если переменная не установлена, используется значение по умолчанию
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
//
// Параметры:
//   - key: имя переменной окружения
//   - defaultValue: значение, которое будет возвращено, если переменная не установлена
//
// Возвращает:
//   - string: значение переменной окружения или defaultValue
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
