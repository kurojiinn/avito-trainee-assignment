package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DB DBConfig
}
type DBConfig struct {
	Host     string `env:"DB_HOST" envDefault:"localhost"`
	Port     string `env:"DB_PORT" envDefault:"5432"`
	Username string `env:"DB_USER" envDefault:"postgres"`
	Password string `env:"DB_PASSWORD" envDefault:"postgres"`
	Name     string `env:"DB_NAME" envDefault:"postgres"`
}

func LoadConfig() *Config {
	var dbConfig DBConfig
	err := envconfig.Process("", &dbConfig)
	if err != nil {
		log.Fatalf("Failed to load DB config: %v", err)
	}
	return &Config{DB: dbConfig}
}
