package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// AdminSeed holds bootstrap credentials for the initial admin user.
type AdminSeed struct {
	Email    string
	Password string
	FullName string
}

// Config centralizes runtime configuration.
type Config struct {
	HTTPPort    string
	DatabaseURL string
	AdminSeed   AdminSeed
	SMTP        SMTPConfig
	JWTSecret   string
	JWTIssuer   string
	JWTTTL      time.Duration
}

// SMTPConfig holds SMTP credentials for email delivery.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	return Config{
		HTTPPort:    envOrDefault("HTTP_PORT", "8080"),
		DatabaseURL: envOrDefault("DATABASE_URL", defaultDatabaseURL()),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		JWTIssuer:   envOrDefault("JWT_ISSUER", "catalog-api"),
		JWTTTL:      durationOrDefault("JWT_TTL", 15*time.Minute),
		SMTP: SMTPConfig{
			Host:     os.Getenv("SMTP_HOST"),
			Port:     intOrDefault("SMTP_PORT", 587),
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     os.Getenv("SMTP_FROM"),
		},
		AdminSeed: AdminSeed{
			Email:    os.Getenv("ADMIN_EMAIL"),
			Password: os.Getenv("ADMIN_PASSWORD"),
			FullName: envOrDefault("ADMIN_FULL_NAME", "Catalog Admin"),
		},
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func defaultDatabaseURL() string {
	user := envOrDefault("POSTGRES_USER", "catalog")
	password := envOrDefault("POSTGRES_PASSWORD", "catalog")
	host := envOrDefault("POSTGRES_HOST", "localhost")
	port := envOrDefault("POSTGRES_PORT", "55432")
	db := envOrDefault("POSTGRES_DB", "catalog")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, db)
}

func durationOrDefault(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if parsed, err := time.ParseDuration(v); err == nil {
			return parsed
		}
	}
	return fallback
}

func intOrDefault(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return fallback
}
