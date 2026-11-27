// Package config loads application settings from environment variables.
package config

import (
	"fmt"
	"os"
)

// Config holds all application settings.
type Config struct {
	ServerPort string
	Database   DatabaseConfig
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// Load reads configuration from the environment, falling back to defaults
// suitable for local development.
func Load() Config {
	return Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "product_management"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}
}

// DSN builds a PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	// timezone=UTC keeps TIMESTAMP columns consistent with the "Z" times in API responses.
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s timezone=UTC",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
