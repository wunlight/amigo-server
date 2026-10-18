package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv       string
	AppPort      string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBSSLMode    string
	DatabaseURL  string
	JWTSecret    string
	JWTExpiry    string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:     env("APP_ENV", "development"),
		AppPort:    env("APP_PORT", "8080"),
		DBHost:     env("DB_HOST", "localhost"),
		DBPort:     env("DB_PORT", "5432"),
		DBUser:     env("DB_USER", "postgres"),
		DBPassword: env("DB_PASSWORD", "postgres"),
		DBName:     env("DB_NAME", "amigo_db"),
		DBSSLMode:  env("DB_SSLMODE", "disable"),
		JWTSecret:  env("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
		JWTExpiry:  env("JWT_EXPIRY", "24h"),
	}

	cfg.DatabaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode,
	)

	return cfg, nil
}

func env(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
