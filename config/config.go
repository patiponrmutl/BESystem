package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppPort string
	AppEnv  string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

func get(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func Load() *Config {
	return &Config{
		AppPort: get("APP_PORT", "8080"),
		AppEnv:  get("APP_ENV", "dev"),

		DBHost:     get("DB_HOST", "localhost"),
		DBPort:     get("DB_PORT", "5432"),
		DBUser:     get("DB_USER", "postgres"),
		DBPassword: get("DB_PASSWORD", "19012546"),
		DBName:     get("DB_NAME", "studentplusadmin"),
		DBSSLMode:  get("DB_SSLMODE", "disable"),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBSSLMode,
	)
}
