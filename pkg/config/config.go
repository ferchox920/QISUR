package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// AdminSeed contiene las credenciales de arranque para el usuario admin inicial.
type AdminSeed struct {
	Email    string
	Password string
	FullName string
}

// Config centraliza la configuracion de runtime.
type Config struct {
	HTTPPort         string
	DatabaseURL      string
	AdminSeed        AdminSeed
	SMTP             SMTPConfig
	JWTSecret        string
	JWTIssuer        string
	JWTTTL           time.Duration
	WSAllowedOrigins []string
}

// SMTPConfig contiene las credenciales SMTP para el envio de correo.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	SkipTLS  bool
}

// Load lee configuracion desde variables de entorno con valores por defecto.
func Load() Config {
	return Config{
		HTTPPort:         envOrDefault("HTTP_PORT", "8080"),
		DatabaseURL:      envOrDefault("DATABASE_URL", defaultDatabaseURL()),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		JWTIssuer:        envOrDefault("JWT_ISSUER", "catalog-api"),
		JWTTTL:           durationOrDefault("JWT_TTL", 15*time.Minute),
		WSAllowedOrigins: splitAndTrim(os.Getenv("WS_ALLOWED_ORIGINS")),
		SMTP: SMTPConfig{
			Host:     os.Getenv("SMTP_HOST"),
			Port:     intOrDefault("SMTP_PORT", 587),
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     os.Getenv("SMTP_FROM"),
			SkipTLS:  boolOrDefault("SMTP_TLS_SKIP_VERIFY", false),
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

func boolOrDefault(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		switch v {
		case "1", "true", "TRUE", "True", "yes", "Y", "y":
			return true
		case "0", "false", "FALSE", "False", "no", "N", "n":
			return false
		}
	}
	return fallback
}

func splitAndTrim(val string) []string {
	if val == "" {
		return nil
	}
	parts := strings.Split(val, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
